package resty

import "net/http"

type RESTHandler interface {
	HandleGet(rw http.ResponseWriter, r *http.Request)
	HandlePost(rw http.ResponseWriter, r *http.Request)
	HandlePut(rw http.ResponseWriter, r *http.Request)
	HandlePatch(rw http.ResponseWriter, r *http.Request)
	HandleDelete(rw http.ResponseWriter, r *http.Request)
	HandleCopy(rw http.ResponseWriter, r *http.Request)
	HandleHead(rw http.ResponseWriter, r *http.Request)
	HandleOptions(rw http.ResponseWriter, r *http.Request)
	HandleLink(rw http.ResponseWriter, r *http.Request)
	HandleUnlink(rw http.ResponseWriter, r *http.Request)
	HandlePurge(rw http.ResponseWriter, r *http.Request)
	HandleLock(rw http.ResponseWriter, r *http.Request)
	HandleUnlock(rw http.ResponseWriter, r *http.Request)
	HandlePropFind(rw http.ResponseWriter, r *http.Request)
	HandleView(rw http.ResponseWriter, r *http.Request)
}

// -----------------------
// default handler. embed this in your handler to provide defaults for all the verbs.

type DefaultRESTHandler struct{}

func (h *DefaultRESTHandler) HandleGet(rw http.ResponseWriter, r *http.Request) {
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
func (h *DefaultRESTHandler) HandleLink(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleUnlink(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePurge(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleLock(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleUnlock(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePropFind(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleView(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}

// ------------------------
// converter middleware to turn a normal handler into a restful one.

func Restful(next RESTHandler) http.Handler {
	return &RestfulConvertMiddleware{NextHandler: next}
}

// UserRestrictMiddleware redirects user to login page if not logged in
type RestfulConvertMiddleware struct {
	NextHandler RESTHandler
}

func (m *RestfulConvertMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		m.NextHandler.HandleGet(rw, r)
	case r.Method == "POST":
		m.NextHandler.HandlePost(rw, r)
	case r.Method == "PUT":
		m.NextHandler.HandlePut(rw, r)
	case r.Method == "PATCH":
		m.NextHandler.HandlePatch(rw, r)
	case r.Method == "DELETE":
		m.NextHandler.HandleDelete(rw, r)
	case r.Method == "COPY":
		m.NextHandler.HandleCopy(rw, r)
	case r.Method == "HEAD":
		m.NextHandler.HandleHead(rw, r)
	case r.Method == "OPTIONS":
		m.NextHandler.HandleOptions(rw, r)
	case r.Method == "LINK":
		m.NextHandler.HandleLink(rw, r)
	case r.Method == "UNLINK":
		m.NextHandler.HandleUnlink(rw, r)
	case r.Method == "PURGE":
		m.NextHandler.HandlePurge(rw, r)
	case r.Method == "LOCK":
		m.NextHandler.HandleLock(rw, r)
	case r.Method == "UNLOCK":
		m.NextHandler.HandleUnlock(rw, r)
	case r.Method == "PROPFIND":
		m.NextHandler.HandlePropFind(rw, r)
	case r.Method == "VIEW":
		m.NextHandler.HandleView(rw, r)
	default:
		http.Error(rw, "Unsupported", http.StatusNotImplemented)
	}
}
