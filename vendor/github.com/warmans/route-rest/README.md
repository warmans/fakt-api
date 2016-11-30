Route Rest
=====================

Takes a configuration and applies a set of restful routes to a gorilla mux router.

### Example

via example/simple/main.go

```go
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
	routes.DefaultRESTCtxHandler
	//since all routes will use the same handler provider a per-instance prefix to be printed to the response
	responsePrefix string
}

//Init is called due to the InitCtx middleware for all requests. If next() is not invoked the request
//will stop here.
func (h *SimpleHandler) Init(rw http.ResponseWriter, r *http.Request, next routes.CtxHandleFunc) {

	//abort the request if the form is unparsable
	if err := r.ParseForm(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	//example of aborting based on some user input e.g. http://localhost:8080/api/foo?abort=1
	if r.Form.Get("abort") != "" {
		http.Error(rw, "aborted", http.StatusOK)
		return
	}

	//invoke the verb handler with extra stuff in the context
	next(rw, r, context.WithValue(context.Background(), "somekey", "somevalue"))

	//always close the request body
	r.Body.Close()
}

//HandleGet is called for GET http://localhost:8080/api/foo/1 <- with ID
func (h *SimpleHandler) HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(rw, "%s:%s:%+v:%s", h.responsePrefix, "GET", mux.Vars(r), ctx.Value("somekey").(string))
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
				routes.InitCtx(&SimpleHandler{responsePrefix: "foo"}), //handler
				[]*routes.Route{ //sub-routes
					routes.NewRoute(
						"bar",
						"{bar_id:[0-9]}",
						routes.InitCtx(&SimpleHandler{responsePrefix: "bar"}),
						[]*routes.Route{},
					),
					routes.NewRoute(
						"baz",
						"{baz_id:[0-9]}",
						routes.InitCtx(&SimpleHandler{responsePrefix: "baz"}),
						[]*routes.Route{},
					),
				},
			),
		},
		[]string{""}, //no prefix on root resource
	)

	//attach it to a http mux with a prefix
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", router))

	//serve the mux
	log.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

```