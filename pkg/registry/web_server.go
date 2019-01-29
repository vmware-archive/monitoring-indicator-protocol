package registry

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pivotal/indicator-protocol/pkg/mtls"
)

type WebServerConfig struct {
	Address       string
	ServerPEMPath string
	ServerKeyPath string
	RootCAPath    string
	*DocumentStore
}

func NewWebServer(c WebServerConfig) (func() error, func() error, error) {
	tlsConfig, err := mtls.NewServerConfig(c.RootCAPath)
	if err != nil {
		return nil, nil, err
	}

	server := &http.Server{
		Addr:    c.Address,
		Handler: newRouter(c.DocumentStore),
		TLSConfig: tlsConfig,
	}

	start := func() error { return server.ListenAndServeTLS(c.ServerPEMPath, c.ServerKeyPath) }
	stop := func() error { return server.Close() }

	return start, stop, nil
}

var httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
	Subsystem: "registry",
	Name:      "http_requests",
	Help:      "The count of http attempts against this registry instance.",
}, []string{"route", "status"})

func init() {
	prometheus.MustRegister(httpRequests)
}

func newRouter(documentStore *DocumentStore) *mux.Router {
	r := mux.NewRouter()
	r.Handle("/metrics", instrumentEndpoint(httpRequests, promhttp.Handler()))
	r.NotFoundHandler = notFound(httpRequests)
	r.HandleFunc("/v1/register", instrumentEndpoint(httpRequests, NewRegisterHandler(documentStore))).Methods(http.MethodPost)
	r.HandleFunc("/v1/indicator-documents", instrumentEndpoint(httpRequests, NewIndicatorDocumentsHandler(documentStore))).Methods(http.MethodGet)
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
