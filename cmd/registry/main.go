package main

import (
	"code.cloudfoundry.org/indicators/pkg/mtls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"code.cloudfoundry.org/indicators/pkg/registry"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
	Subsystem: "registry",
	Name:      "http_requests",
	Help:      "The count of http attempts against this registry instance.",
}, []string{"route", "status"})

func init() {
	prometheus.MustRegister(httpRequests)
}

func main() {
	port := flag.Int("port", -1, "Port to expose registration endpoints")
	serverPEM := flag.String("tls-pem-path", "", "Server TLS public cert pem path")
	serverKey := flag.String("tls-key-path", "", "Server TLS private key path")
	rootCACert := flag.String("tls-root-ca-pem", "", "Root CA Pem for self-signed certs.")
	expiration := flag.Duration("indicator-expiration", 120*time.Minute, "Document expiration duration")
	flag.Parse()

	address := fmt.Sprintf(":%d", *port)

	start, stop, err := mtls.NewServer(address, *serverPEM, *serverKey, *rootCACert, newRouter(*expiration))
	defer stop()
	if err != nil {
		log.Fatalf("failed to create server: %s\n", err)
	}

	err = start()
	if err != nil {
		log.Fatalf("failed to create server: %s\n", err)
	}
}

func newRouter(indicatorExpiration time.Duration) *mux.Router {
	documentStore := registry.NewDocumentStore(indicatorExpiration)

	r := mux.NewRouter()
	r.Handle("/metrics", instrumentEndpoint(httpRequests, promhttp.Handler()))
	r.NotFoundHandler = notFound(httpRequests)
	r.HandleFunc("/v1/register", instrumentEndpoint(httpRequests, registry.NewRegisterHandler(documentStore))).Methods(http.MethodPost)
	r.HandleFunc("/v1/indicator-documents", instrumentEndpoint(httpRequests, registry.NewIndicatorDocumentsHandler(documentStore))).Methods(http.MethodGet)
	return r
}

func notFound(counter *prometheus.CounterVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("404 page not found"))

		log.Println("404 returned for path: ", r.URL.Path)
		counter.WithLabelValues("invalid path", "404").Inc()
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(statusCode int) {
	sr.status = statusCode
	sr.ResponseWriter.WriteHeader(statusCode)
}

func instrumentEndpoint(counter *prometheus.CounterVec, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := statusRecorder{ResponseWriter: w, status: 200}

		h.ServeHTTP(&rec, r)

		counter.WithLabelValues(r.URL.Path, strconv.Itoa(rec.status)).Inc()
	}
}
