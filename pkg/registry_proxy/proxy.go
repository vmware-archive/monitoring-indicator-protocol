package registry_proxy

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	localRegistryHandler http.Handler
	registryHandlers     []http.Handler
}

type noopWriter struct{}

func (*noopWriter) Header() http.Header {
	return make(http.Header)
}

func (*noopWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (*noopWriter) WriteHeader(int) {}

var defaultNoopWriter = &noopWriter{}

func NewHandler(localRegistryHandler http.Handler, registryHandlers []http.Handler) *Handler {
	return &Handler{
		localRegistryHandler: localRegistryHandler,
		registryHandlers:     registryHandlers,
	}
}

var client = &http.Client{
	Timeout: 15 * time.Second,
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	const prefix = "/backend/"

	if strings.HasPrefix(r.URL.Path, prefix) {
		r.URL.Path = r.URL.Path[len(prefix)-1:]
		h.localRegistryHandler.ServeHTTP(rw, r)
		return
	}

	if r.Method != http.MethodPost {
		h.localRegistryHandler.ServeHTTP(rw, r)
		return
	}
	buff := new(bytes.Buffer)
	_, err := buff.ReadFrom(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	requestBody := buff.String()

	newReq, err := copyRequest(r, requestBody)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.localRegistryHandler.ServeHTTP(rw, newReq)
	r.URL.Path = prefix[:len(prefix)-1] + r.URL.Path

	var wg sync.WaitGroup
	wg.Add(len(h.registryHandlers))

	for _, handler := range h.registryHandlers {
		newReq, err := copyRequest(r, requestBody)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		go func(currentHandler http.Handler) {
			defer wg.Done()
			currentHandler.ServeHTTP(defaultNoopWriter, newReq)
		}(handler)
	}
	wg.Wait()
}

func copyRequest(r *http.Request, requestBody string) (*http.Request, error) {
	newReq, err := http.NewRequest(r.Method, r.URL.String(), ioutil.NopCloser(bytes.NewReader([]byte(requestBody))))
	if err != nil {
		return nil, errors.New("could not copy request to send to proxy")
	}
	newReq.Header = r.Header
	newReq.Host = r.Host
	newReq.ContentLength = r.ContentLength
	return newReq, nil
}
