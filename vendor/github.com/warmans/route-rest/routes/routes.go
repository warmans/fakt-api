package routes

import (
	"strings"

	"fmt"

	"github.com/gorilla/mux"
)

type Route struct {
	Name      string
	IDPattern string
	Handler   RESTHandler
	Sub       []*Route
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

		uriHandlers := make([]*mux.Route, 0)

		//route without ID
		cget := append(parentURI, route.Name)

		//route with ID
		get := append(cget, route.IDPattern)

		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), route.Handler.HandleGetList).Methods("GET").Name(fmt.Sprintf("%s:%s", route.Name, "cget")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), route.Handler.HandleGet).Methods("GET").Name(fmt.Sprintf("%s:%s", route.Name, "get")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), route.Handler.HandlePost).Methods("POST").Name(fmt.Sprintf("%s:%s", route.Name, "post")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), route.Handler.HandlePut).Methods("PUT").Name(fmt.Sprintf("%s:%s", route.Name, "put")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), route.Handler.HandleDelete).Methods("DELETE").Name(fmt.Sprintf("%s:%s", route.Name, "delete")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), route.Handler.HandlePatch).Methods("PATCH").Name(fmt.Sprintf("%s:%s", route.Name, "patch")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(get, "/"), route.Handler.HandleCopy).Methods("COPY").Name(fmt.Sprintf("%s:%s", route.Name, "copy")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), route.Handler.HandleHead).Methods("HEAD").Name(fmt.Sprintf("%s:%s", route.Name, "head")),
		)
		uriHandlers = append(
			uriHandlers,
			router.HandleFunc(strings.Join(cget, "/"), route.Handler.HandleOptions).Methods("OPTIONS").Name(fmt.Sprintf("%s:%s", route.Name, "options")),
		)

		if route.Sub != nil {
			ApplyRoutes(router, route.Sub, get)
		}

	}
}
