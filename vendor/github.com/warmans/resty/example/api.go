package main

// to run this file just `go run api.go`

import (
	"net/http"

	"fmt"
	"log"

	"github.com/warmans/resty"
)

type SomeResource struct {
	resty.DefaultRESTHandler
}

func (h *SomeResource) HandleGet(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "GET")
}
func (h *SomeResource) HandlePost(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "POST")
}

func main() {

	mux := http.NewServeMux()
	mux.Handle("/resource", resty.Restful(&SomeResource{}))

	log.Print("Listening on localhost:8080. Try and send a POST or GET")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
