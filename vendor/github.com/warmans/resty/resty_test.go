package resty

import (
	"net/http"
	"testing"

	"bytes"
	"fmt"
	"net/http/httptest"

	"strings"
)

type TestRestHandler struct {
	DefaultRESTHandler
}

func (h *TestRestHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "GET")
}
func (h *TestRestHandler) HandlePost(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "POST")
}
func (h *TestRestHandler) HandlePut(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "PUT")
}
func (h *TestRestHandler) HandlePatch(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "PATCH")
}
func (h *TestRestHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "DELETE")
}
func (h *TestRestHandler) HandleCopy(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "COPY")
}
func (h *TestRestHandler) HandleHead(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "HEAD")
}
func (h *TestRestHandler) HandleOptions(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "OPTIONS")
}
func (h *TestRestHandler) HandleLink(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "LINK")
}
func (h *TestRestHandler) HandleUnlink(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "UNLINK")
}
func (h *TestRestHandler) HandlePurge(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "PURGE")
}
func (h *TestRestHandler) HandleLock(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "LOCK")
}
func (h *TestRestHandler) HandleUnlock(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "UNLOCK")
}
func (h *TestRestHandler) HandlePropFind(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "PROPFIND")
}
func (h *TestRestHandler) HandleView(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprint(rw, "VIEW")
}

func TestRestfulConvertMiddleware(t *testing.T) {

	tests := []struct {
		Method         string
		ExpectedBody   string
		ExpectedStatus int
	}{
		{Method: "GET", ExpectedBody: "GET", ExpectedStatus: http.StatusOK},
		{Method: "POST", ExpectedBody: "POST", ExpectedStatus: http.StatusOK},
		{Method: "PUT", ExpectedBody: "PUT", ExpectedStatus: http.StatusOK},
		{Method: "PATCH", ExpectedBody: "PATCH", ExpectedStatus: http.StatusOK},
		{Method: "DELETE", ExpectedBody: "DELETE", ExpectedStatus: http.StatusOK},
		{Method: "COPY", ExpectedBody: "COPY", ExpectedStatus: http.StatusOK},
		{Method: "HEAD", ExpectedBody: "HEAD", ExpectedStatus: http.StatusOK},
		{Method: "OPTIONS", ExpectedBody: "OPTIONS", ExpectedStatus: http.StatusOK},
		{Method: "LINK", ExpectedBody: "LINK", ExpectedStatus: http.StatusOK},
		{Method: "UNLINK", ExpectedBody: "UNLINK", ExpectedStatus: http.StatusOK},
		{Method: "PURGE", ExpectedBody: "PURGE", ExpectedStatus: http.StatusOK},
		{Method: "LOCK", ExpectedBody: "LOCK", ExpectedStatus: http.StatusOK},
		{Method: "UNLOCK", ExpectedBody: "UNLOCK", ExpectedStatus: http.StatusOK},
		{Method: "PROPFIND", ExpectedBody: "PROPFIND", ExpectedStatus: http.StatusOK},
		{Method: "VIEW", ExpectedBody: "VIEW", ExpectedStatus: http.StatusOK},
		{Method: "FOO", ExpectedBody: "Unsupported", ExpectedStatus: http.StatusNotImplemented},
	}

	handler := &TestRestHandler{}

	for _, test := range tests {

		rw := httptest.NewRecorder()
		r, err := http.NewRequest(test.Method, "", bytes.NewBuffer([]byte("")))

		if err != nil {
			t.Errorf("Error building test request: %s", err)
		}

		//execute the handler
		Restful(handler).ServeHTTP(rw, r)

		if res := strings.TrimSpace(rw.Body.String()); res != test.ExpectedBody {
			t.Errorf("Unexpected response body! Expected '%s' got '%s'", test.ExpectedBody, res)
		}
		if rw.Code != test.ExpectedStatus {
			t.Errorf("Unexpected status! Expected %d got %d", test.ExpectedStatus, rw.Code)
		}
	}
}

func TestDefaultRESTHandler(t *testing.T) {

	handler := &DefaultRESTHandler{}

	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "LINK", "UNLINK", "PURGE", "LOCK", "UNLOCK", "PROPFIND", "VIEW"} {

		rw := httptest.NewRecorder()

		r, err := http.NewRequest(method, "", bytes.NewBuffer([]byte("")))
		if err != nil {
			t.Errorf("Error building test request: %s", err)
		}

		//execute the handler
		Restful(handler).ServeHTTP(rw, r)

		if res := strings.TrimSpace(rw.Body.String()); res != "Not Implemented" {
			t.Errorf("Unexpected response body!Default should always return 'Not Implemented' actually: '%s'", res)
		}

		if rw.Code != http.StatusNotImplemented {
			t.Errorf("Unexpected status! Expected %d got %d", http.StatusNotImplemented, rw.Code)
		}
	}
}
