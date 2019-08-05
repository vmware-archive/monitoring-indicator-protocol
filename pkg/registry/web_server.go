package registry

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pivotal/monitoring-indicator-protocol/pkg/registry/status_store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type WebServerConfig struct {
	Address       string
	DocumentStore *DocumentStore
	StatusStore   *status_store.Store
}

func NewWebServer(c WebServerConfig) (func() error, func() error) {
	server := &http.Server{
		Addr:         c.Address,
		Handler:      newRouter(c),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	start := func() error { return server.ListenAndServe() }
	stop := func() error { return server.Close() }

	return start, stop
}

var httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
	Subsystem: "registry",
	Name:      "http_requests",
	Help:      "The count of http attempts against this registry instance.",
}, []string{"route", "status"})

func init() {
	prometheus.MustRegister(httpRequests)
}

func newRouter(w WebServerConfig) *mux.Router {
	r := mux.NewRouter()
	r.Handle("/metrics", instrumentEndpoint(httpRequests, promhttp.Handler()))
	r.NotFoundHandler = notFound(httpRequests)

	// Optional trailing slash: https://github.com/gorilla/mux/issues/30#issuecomment-321045004
	optionalTrailingSlash := "{_:(?:\\/)?}"
	r.HandleFunc("/v1/register" + optionalTrailingSlash,
		instrumentEndpoint(httpRequests, NewRegisterHandler(w.DocumentStore))).Methods(http.MethodPost)
	r.HandleFunc("/v1/indicator-documents" + optionalTrailingSlash,
		instrumentEndpoint(httpRequests, NewIndicatorDocumentsHandler(w.DocumentStore, w.StatusStore))).Methods(http.MethodGet)
	r.HandleFunc("/v1/indicator-documents/{documentID}/bulk_status" + optionalTrailingSlash,
		instrumentEndpoint(httpRequests, NewIndicatorStatusBulkUpdateHandler(w.StatusStore))).Methods(http.MethodPost)
	return r
}

func notFound(counter *prometheus.CounterVec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("404 page not found"))

		log.Println("404 returned")
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

		urlLabel := r.URL.Path
		if strings.Contains(r.URL.Path, "bulk_status") {
			urlLabel = "/v1/indicator-documents/bulk_status"
		}

		counter.WithLabelValues(urlLabel, strconv.Itoa(rec.status)).Inc()
	}
}
