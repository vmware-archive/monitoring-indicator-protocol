package main

import (
	"crypto/tls"
	"log"

	envstruct "code.cloudfoundry.org/go-envstruct"
	"github.com/pivotal/monitoring-indicator-protocol/k8s/pkg/webhook"
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
		log.Fatalf("Failed to load config from environment: %s", err)
	}
	if err := envstruct.WriteReport(&cfg); err != nil {
		log.Printf("Unable to write envstruct report: %s", err)
	}

	cert, err := tls.LoadX509KeyPair(cfg.Cert, cfg.Key)
	if err != nil {
		log.Fatalf("Unable to load certs: %s", err)
	}
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	webhook.NewServer(cfg.HTTPAddr, webhook.WithTLSConfig(tlsConf)).Run(true)
}
