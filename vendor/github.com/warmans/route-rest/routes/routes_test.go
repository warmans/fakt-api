package routes

import (
	"net/http"
	"testing"

	"bytes"
	"fmt"
	"io/ioutil"
	"net/http/httptest"

	"log"
)

type EmptyHandler struct{ DefaultRESTHandler }

func (h *EmptyHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "GET")
}
func (h *EmptyHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "GET LIST")
}
func (h *EmptyHandler) HandlePost(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "POST")
}
func (h *EmptyHandler) HandlePut(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "PUT")
}
func (h *EmptyHandler) HandlePatch(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "PATCH")
}
func (h *EmptyHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "DELETE")
}
func (h *EmptyHandler) HandleCopy(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "COPY")
}
func (h *EmptyHandler) HandleHead(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "HEAD")
}
func (h *EmptyHandler) HandleOptions(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%s", "OPTIONS")
}

func eh() RESTHandler {
	return &EmptyHandler{}
}

func MustNewRequest(method, urlStr string) *http.Request {
	req, err := http.NewRequest(method, urlStr, bytes.NewBufferString(""))
	if err != nil {
		log.Fatalf("Unexpected error: %s", err.Error())
	}
	return req
}

func TestGetRoutes(t *testing.T) {

	testData := map[string]struct {
		routes            []*Route
		expectNamedRoutes map[string]string
	}{
		"Single resource": {
			routes: []*Route{NewRoute("foo", "{id:[0-9]}", eh(), nil)},
			expectNamedRoutes: map[string]string{
				"foo:cget": "/foo",
				"foo:get":  "/foo/{id:[0-9]}",
			},
		},
		"Sub resource": {
			routes: []*Route{NewRoute("foo", "{foo_id:[0-9]}", eh(), []*Route{NewRoute("bar", "{bar_id:[0-9]}", eh(), nil)})},
			expectNamedRoutes: map[string]string{
				"foo:cget": "/foo",
				"foo:get":  "/foo/{foo_id:[0-9]}",
				"bar:cget": "/foo/{foo_id:[0-9]}/bar",
				"bar:get":  "/foo/{foo_id:[0-9]}/bar/{bar_id:[0-9]}",
			},
		},
		"Sub sub resource": {
			routes: []*Route{NewRoute("foo", "{foo_id:[0-9]}", eh(), []*Route{NewRoute("bar", "{bar_id:[0-9]}", eh(), []*Route{NewRoute("baz", "{baz_id:[0-9]}", eh(), nil)})})},
			expectNamedRoutes: map[string]string{
				"foo:cget": "/foo",
				"foo:get":  "/foo/{foo_id:[0-9]}",
				"bar:cget": "/foo/{foo_id:[0-9]}/bar",
				"bar:get":  "/foo/{foo_id:[0-9]}/bar/{bar_id:[0-9]}",
				"baz:cget": "/foo/{foo_id:[0-9]}/bar/{bar_id:[0-9]}/baz",
				"baz:get":  "/foo/{foo_id:[0-9]}/bar/{bar_id:[0-9]}/baz/{baz_id:[0-9]}",
			},
		},
		"Test other verbs": {
			routes: []*Route{NewRoute("foo", "{id:[0-9]}", eh(), nil)},
			expectNamedRoutes: map[string]string{
				"foo:post":    "/foo",
				"foo:head":    "/foo",
				"foo:options": "/foo",
				"foo:put":     "/foo/{id:[0-9]}",
				"foo:patch":   "/foo/{id:[0-9]}",
				"foo:delete":  "/foo/{id:[0-9]}",
				"foo:copy":    "/foo/{id:[0-9]}",
			},
		},
	}

	for testName, test := range testData {

		router := GetRouter(test.routes, []string{""})

		for expectedNamedRoute, expectValidRoute := range test.expectNamedRoutes {
			route := router.GetRoute(expectedNamedRoute)
			if route == nil {
				t.Errorf("%s: failed to find named route: %s", testName, expectedNamedRoute)
				continue
			}
			if tp, _ := route.GetPathTemplate(); tp != expectValidRoute {
				t.Errorf("%s: named route (%s) didn't have correct template. Expected %s got %s", testName, expectedNamedRoute, expectValidRoute, tp)
				continue
			}
		}

	}
}

func TestGetRoutesIntegration(t *testing.T) {

	router := GetRouter(
		[]*Route{
			NewRoute(
				"foo",
				"{foo_id:[0-9]}",
				eh(),
				[]*Route{
					NewRoute(
						"bar",
						"{bar_id:[0-9]}",
						eh(),
						[]*Route{
							NewRoute(
								"baz",
								"{baz_id:[0-9]}",
								eh(),
								nil,
							),
						},
					),
				},
			),
		},
		[]string{""},
	)

	srv := httptest.NewServer(router)
	defer srv.Close()

	testRequests := []struct {
		request          *http.Request
		expectStatusCode int
		expectResponse   string
	}{
		{request: MustNewRequest("GET", fmt.Sprintf("%s%s", srv.URL, "/foo")), expectResponse: "GET LIST", expectStatusCode: http.StatusOK},
		{request: MustNewRequest("GET", fmt.Sprintf("%s%s", srv.URL, "/foo/1")), expectResponse: "GET", expectStatusCode: http.StatusOK},
		{request: MustNewRequest("POST", fmt.Sprintf("%s%s", srv.URL, "/foo")), expectResponse: "POST", expectStatusCode: http.StatusOK},
		{request: MustNewRequest("PUT", fmt.Sprintf("%s%s", srv.URL, "/foo/1")), expectResponse: "PUT", expectStatusCode: http.StatusOK},
		{request: MustNewRequest("PATCH", fmt.Sprintf("%s%s", srv.URL, "/foo/1")), expectResponse: "PATCH", expectStatusCode: http.StatusOK},
		{request: MustNewRequest("DELETE", fmt.Sprintf("%s%s", srv.URL, "/foo/1")), expectResponse: "DELETE", expectStatusCode: http.StatusOK},
		{request: MustNewRequest("POST", fmt.Sprintf("%s%s", srv.URL, "/foo/1")), expectResponse: "", expectStatusCode: http.StatusNotFound},
	}

	for _, test := range testRequests {
		response, err := http.DefaultClient.Do(test.request)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
			return
		}

		b, err := ioutil.ReadAll(response.Body)

		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
			return
		}

		if response.StatusCode != test.expectStatusCode {
			t.Errorf("Unexpected status. Expected %d got %d", test.expectStatusCode, response.StatusCode)
			return
		}

		if test.expectStatusCode == http.StatusOK {
			if string(b) != test.expectResponse {
				t.Errorf("Body did not contain expected data. Expected %s got %s", test.expectResponse, string(b))
				return
			}
		}
	}
}
