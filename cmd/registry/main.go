package main

import (
	"flag"
	"net/http"
	"fmt"
	"strconv"

	"code.cloudfoundry.org/cf-indicators/pkg/registry"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
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
	flag.Parse()

	documentStore := registry.NewDocumentStore()

	r := mux.NewRouter()

	r.Handle("/metrics", MetricTrack(httpRequests, promhttp.Handler()))
	r.NotFoundHandler = notFound(httpRequests)

	r.HandleFunc("/v1/register", MetricTrack(httpRequests, registry.NewRegisterHandler(documentStore))).Methods(http.MethodPost)
	r.HandleFunc("/v1/indicator-documents", MetricTrack(httpRequests, registry.NewIndicatorDocumentsHandler(documentStore))).Methods(http.MethodGet)

	http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
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

func MetricTrack(counter *prometheus.CounterVec, h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rec := statusRecorder{ResponseWriter: w, status: 200}

		h.ServeHTTP(&rec, r)

		counter.WithLabelValues(r.URL.Path, strconv.Itoa(rec.status)).Inc()
	}
}
