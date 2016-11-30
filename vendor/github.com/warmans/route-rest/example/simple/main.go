package main

import (
	"fmt"
	"net/http"

	"log"

	"github.com/gorilla/mux"
	"github.com/warmans/route-rest/routes"
	"golang.org/x/net/context"
)

//Simple handler just implements a few verbs and prints some data from the request/context
type SimpleHandler struct {
	//include default handlers so the struct implements RESTCtxHandler
	routes.DefaultRESTHandler
	//since all routes will use the same handler provider a per-instance prefix to be printed to the response
	responsePrefix string
}

//HandleGet is called for GET http://localhost:8080/api/foo/1 <- with ID
func (h *SimpleHandler) HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "%s:%s:%+v", h.responsePrefix, "GET", mux.Vars(r))
}

//HandleGetList is called for GET http://localhost:8080/api/foo <-- without ID
func (h *SimpleHandler) HandleGetList(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "%s:%s:%+v", h.responsePrefix, "GET LIST", mux.Vars(r))
}

//HandleGetList is called for POST http://localhost:8080/api/foo <-- without ID
func (h *SimpleHandler) HandlePost(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "%s:%s:%+v", h.responsePrefix, "POST", mux.Vars(r))
}

func main() {

	//start with a new muxRouter
	router := mux.NewRouter()

	//apply the REST routes. This will create the following routes:
	// /foo
	// /foo/1
	// /foo/1/bar
	// /foo/1/bar/1
	// /foo/1/baz
	// /foo/1/baz/1
	routes.ApplyRoutes(
		router,
		[]*routes.Route{
			routes.NewRoute(
				"foo",                                                 //resource name
				"{foo_id:[0-9]}",                                      //id pattern
				&SimpleHandler{responsePrefix: "foo"}, //handler
				[]*routes.Route{ //sub-routes
					routes.NewRoute(
						"bar",
						"{bar_id:[0-9]}",
						&SimpleHandler{responsePrefix: "bar"},
						[]*routes.Route{},
					),
					routes.NewRoute(
						"baz",
						"{baz_id:[0-9]}",
						&SimpleHandler{responsePrefix: "baz"},
						[]*routes.Route{},
					),
				},
			),
		},
		[]string{""}, //no prefix on root resource
	)

	//attach it to a http mux with a prefix
	muxer := http.NewServeMux()
	muxer.Handle("/api/", http.StripPrefix("/api", router))

	//serve the mux
	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", muxer))
}
