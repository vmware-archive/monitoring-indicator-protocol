package main

import (
	"flag"
	"net/http"
	"fmt"

	"github.com/gorilla/mux"
	"code.cloudfoundry.org/cf-indicators/pkg/registry"
)

func main() {
	port := flag.Int("port", -1, "Port to expose registration endpoints")
	flag.Parse()

	documentStore := registry.NewDocumentStore()

	r := mux.NewRouter()
	// TODO set methods
	r.HandleFunc("/v1/register", registry.NewRegisterHandler(documentStore))
	r.HandleFunc("/v1/indicator-documents", registry.NewIndicatorDocumentsHandler(documentStore))

	http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
}
