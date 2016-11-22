package main

// to run this file just `go run api.go`

import (
	"net/http"

	"fmt"
	"log"

	"github.com/warmans/resty"
	"golang.org/x/net/context"
	"github.com/warmans/ctxhandler"
)

type SomeResource struct {
	resty.DefaultRESTHandler
}

func (h *SomeResource) Before(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "ALWAYS -> ")
}
func (h *SomeResource) HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "GET")
}
func (h *SomeResource) HandlePost(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "POST")
}

func main() {

	mux := http.NewServeMux()

	//add functionality: resty.RESTHandler -> resty.CtxHandler -> http.Handler
	mux.Handle("/resource", ctxhandler.Ctx(resty.Restful(&SomeResource{})))

	log.Printf("Listening on localhost:8080. Try and send a POST or GET")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
