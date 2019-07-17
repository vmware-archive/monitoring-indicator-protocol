package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client/clientset/versioned"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/lifecycle"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"code.cloudfoundry.org/go-envstruct"
	informers "github.com/pivotal/monitoring-indicator-protocol/pkg/k8s/client/informers/externalversions"
)

type config struct {
	Namespace string `env:"NAMESPACE,required,report"`
}

func init() {
	klog.SetOutput(os.Stderr)
}

func main() {
	flag.Parse()
	ctx := setupSignalHandler()

	var conf config
	err := envstruct.Load(&conf)
	if err != nil {
		log.Fatal("failed to load env variables: NAMESPACE is required")
	}
	err = envstruct.WriteReport(&conf)
	if err != nil {
		log.Fatal("failed to write report using env variables")
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("failed to configure kubernetes cluster; make sure kubernetes is running")
	}

	client, err := versioned.NewForConfig(cfg)
	if err != nil {
		log.Fatal("failed to create clientSet for the given config")
	}

	controller := lifecycle.NewController(
		client.AppsV1alpha1(),
	)

	informerFactory := informers.NewSharedInformerFactory(
		client,
		time.Second*30,
	)

	indicatorInformer := informerFactory.Apps().
		V1alpha1().
		IndicatorDocuments().
		Informer()
	indicatorInformer.AddEventHandler(controller)

	log.Println("running informer...")
	indicatorInformer.Run(ctx.Done())
}

var onlyOneSignalHandler = make(chan struct{})

// setupSignalHandler registers SIGTERM and SIGINT. A context is returned
// which is canceled on one of these signals. If a second signal is caught,
// the program is terminated with exit code 1.
func setupSignalHandler() context.Context {
	close(onlyOneSignalHandler) // only call once, panic on calls > 1

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cancel()
		<-c
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
