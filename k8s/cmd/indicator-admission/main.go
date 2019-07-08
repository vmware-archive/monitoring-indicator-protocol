package main

import (
	"crypto/tls"
	"log"

	"code.cloudfoundry.org/go-envstruct"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/admission"
)

type config struct {
	HTTPAddr string `env:"HTTP_ADDR, required, report"`
	Cert     string `env:"INDICATOR_ADMISSION_CERT, required, report"`
	Key      string `env:"INDICATOR_ADMISSION_KEY, required, report"`
}

func main() {
	cfg := config{
		Cert: "/etc/indicator-admission-certs/tls.crt",
		Key:  "/etc/indicator-admission-certs/tls.key",
	}
	if err := envstruct.Load(&cfg); err != nil {
		log.Fatal("Failed to load config from environment")
	}
	if err := envstruct.WriteReport(&cfg); err != nil {
		log.Print("Unable to write envstruct report")
	}

	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
	if err != nil {
		log.Fatal("Unable to load certs")
	}
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	admission.NewServer(cfg.HTTPAddr, admission.WithTLSConfig(tlsConf)).Run(true)
}
