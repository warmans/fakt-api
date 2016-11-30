package routes

import (
	"net/http"
)

type CtxHandleFunc func(rw http.ResponseWriter, r *http.Request)

type RESTHandler interface {
	HandleGet(rw http.ResponseWriter, r *http.Request)
	HandleGetList(rw http.ResponseWriter, r *http.Request)
	HandlePost(rw http.ResponseWriter, r *http.Request)
	HandlePut(rw http.ResponseWriter, r *http.Request)
	HandlePatch(rw http.ResponseWriter, r *http.Request)
	HandleDelete(rw http.ResponseWriter, r *http.Request)
	HandleCopy(rw http.ResponseWriter, r *http.Request)
	HandleHead(rw http.ResponseWriter, r *http.Request)
	HandleOptions(rw http.ResponseWriter, r *http.Request)
}

// -----------------------
// default handler. embed this in a handler to follow the RESTCtxHandler interface without needing to implement all the
// verbs

type DefaultRESTHandler struct{}

func (h *DefaultRESTHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePost(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePut(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePatch(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleCopy(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleHead(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleOptions(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}