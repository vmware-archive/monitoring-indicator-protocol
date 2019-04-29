package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/clientset/versioned"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/indicator_status"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/prometheus_uaa_client"
	"k8s.io/client-go/rest"
	"k8s.io/klog"

	"code.cloudfoundry.org/go-envstruct"
	informers "github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/client/informers/externalversions"
)

type config struct {
	Namespace          string `env:"NAMESPACE,required,report"`
	PrometheusURL      string `env:"PROMETHEUS_URL,required,report"`
	PrometheusApiToken string `env:"PROMETHEUS_API_TOKEN,report"`
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
		log.Fatal(err.Error())
	}
	err = envstruct.WriteReport(&conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg, err := rest.InClusterConfig()
	cfg.Timeout = 5*time.Second
	if err != nil {
		log.Fatal(err.Error())
	}

	client, err := versioned.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	tokenFetcher := func() (string, error) { return conf.PrometheusApiToken, nil }
	prometheusClient, err := prometheus_uaa_client.Build(conf.PrometheusURL, tokenFetcher, false)
	if err != nil {
		log.Fatal(err.Error())
	}

	controller := indicator_status.NewController(
		client.AppsV1alpha1(),
		prometheusClient,
		30*time.Second,
		clock.New(),
		conf.Namespace,
	)

	informerFactory := informers.NewSharedInformerFactory(
		client,
		time.Second*30,
	)

	indicatorInformer := informerFactory.Apps().
		V1alpha1().
		Indicators().
		Informer()
	indicatorInformer.AddEventHandler(controller)

	go controller.Start()

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
