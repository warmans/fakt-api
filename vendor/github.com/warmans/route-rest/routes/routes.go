package routes

import (
	"strings"

	"fmt"

	"github.com/gorilla/mux"
	"net/http"
)

type Middleware func (next http.HandlerFunc) http.HandlerFunc

func addNoopMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(rw, r)
	}
}

type Route struct {
	Name      string
	IDPattern string
	Handler   RESTHandler
	Sub       []*Route
	Mw        Middleware
}

func (r *Route) Middleware(mw Middleware) *Route {
	r.Mw = mw
	return r
}

func NewRoute(name string, idPattern string, handler RESTHandler, sub []*Route) *Route {
	return &Route{Name: name, IDPattern: idPattern, Handler: handler, Sub: sub}
}

func GetRouter(routes []*Route, parentURI []string) *mux.Router {
	router := mux.NewRouter()
	ApplyRoutes(router, routes, parentURI)
	return router
}

func ApplyRoutes(router *mux.Router, routes []*Route, parentURI []string) {

	for _, route := range routes {

		//this is a bit weird. To apply middleware a function must be passed that implements Middleware. this function
		//takes a handler func (which will be the method on the RESTHandler) and wraps it in another handler func.
		//multiple middlewares can be nested in this way.
		//if none is used a null middleware is used that just passes through.
		var mw Middleware
		if route.Mw != nil {
			mw = route.Mw
		} else {
			mw = addNoopMiddleware
		}

		uriHandlers := make([]*mux.Route, 0)

		//route without ID
		cget := append(parentURI, route.Name)

		//route with ID
		get := append(cget, route.IDPattern)

		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), mw(route.Handler.HandleGetList).ServeHTTP).Methods("GET").Name(fmt.Sprintf("%s:%s", route.Name, "cget")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), mw(route.Handler.HandleGet).ServeHTTP).Methods("GET").Name(fmt.Sprintf("%s:%s", route.Name, "get")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), mw(route.Handler.HandlePost).ServeHTTP).Methods("POST").Name(fmt.Sprintf("%s:%s", route.Name, "post")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), mw(route.Handler.HandlePut).ServeHTTP).Methods("PUT").Name(fmt.Sprintf("%s:%s", route.Name, "put")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), mw(route.Handler.HandleDelete).ServeHTTP).Methods("DELETE").Name(fmt.Sprintf("%s:%s", route.Name, "delete")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), mw(route.Handler.HandlePatch).ServeHTTP).Methods("PATCH").Name(fmt.Sprintf("%s:%s", route.Name, "patch")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), mw(route.Handler.HandleCopy).ServeHTTP).Methods("COPY").Name(fmt.Sprintf("%s:%s", route.Name, "copy")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), mw(route.Handler.HandleHead).ServeHTTP).Methods("HEAD").Name(fmt.Sprintf("%s:%s", route.Name, "head")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), mw(route.Handler.HandleOptions).ServeHTTP).Methods("OPTIONS").Name(fmt.Sprintf("%s:%s", route.Name, "options")),
		)

		if route.Sub != nil {
			ApplyRoutes(router, route.Sub, get)
		}

	}
}
