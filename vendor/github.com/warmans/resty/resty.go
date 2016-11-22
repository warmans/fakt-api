package resty

import (
	"net/http"

	"golang.org/x/net/context"
	"github.com/warmans/ctxhandler"
)

type RESTHandler interface {
	Before(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandlePost(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandlePut(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandlePatch(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleDelete(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleCopy(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleHead(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleOptions(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleLink(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleUnlink(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandlePurge(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleLock(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleUnlock(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandlePropFind(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	HandleView(rw http.ResponseWriter, r *http.Request, ctx context.Context)
	After(rw http.ResponseWriter, r *http.Request, ctx context.Context)
}

// -----------------------
// default handler. embed this in your handler to provide defaults for all the verbs.

type DefaultRESTHandler struct{}

func (h *DefaultRESTHandler) Before(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	//do nothing
}
func (h *DefaultRESTHandler) HandleGet(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePost(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePut(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePatch(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleDelete(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleCopy(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleHead(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleOptions(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleLink(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleUnlink(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePurge(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleLock(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleUnlock(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandlePropFind(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) HandleView(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	http.Error(rw, "Not Implemented", http.StatusNotImplemented)
}
func (h *DefaultRESTHandler) After(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	//do nothing
}

// ------------------------
// converter middleware to turn a normal handler into a restful one.

func Restful(next RESTHandler) ctxhandler.CtxHandler {
	return &RestfulConvertMiddleware{NextHandler: next}
}

// UserRestrictMiddleware redirects user to login page if not logged in
type RestfulConvertMiddleware struct {
	NextHandler RESTHandler
}

func (m *RestfulConvertMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	m.NextHandler.Before(rw, r, ctx)
	switch {
	case r.Method == "GET":
		m.NextHandler.HandleGet(rw, r, ctx)
	case r.Method == "POST":
		m.NextHandler.HandlePost(rw, r, ctx)
	case r.Method == "PUT":
		m.NextHandler.HandlePut(rw, r, ctx)
	case r.Method == "PATCH":
		m.NextHandler.HandlePatch(rw, r, ctx)
	case r.Method == "DELETE":
		m.NextHandler.HandleDelete(rw, r, ctx)
	case r.Method == "COPY":
		m.NextHandler.HandleCopy(rw, r, ctx)
	case r.Method == "HEAD":
		m.NextHandler.HandleHead(rw, r, ctx)
	case r.Method == "OPTIONS":
		m.NextHandler.HandleOptions(rw, r, ctx)
	case r.Method == "LINK":
		m.NextHandler.HandleLink(rw, r, ctx)
	case r.Method == "UNLINK":
		m.NextHandler.HandleUnlink(rw, r, ctx)
	case r.Method == "PURGE":
		m.NextHandler.HandlePurge(rw, r, ctx)
	case r.Method == "LOCK":
		m.NextHandler.HandleLock(rw, r, ctx)
	case r.Method == "UNLOCK":
		m.NextHandler.HandleUnlock(rw, r, ctx)
	case r.Method == "PROPFIND":
		m.NextHandler.HandlePropFind(rw, r, ctx)
	case r.Method == "VIEW":
		m.NextHandler.HandleView(rw, r, ctx)
	default:
		http.Error(rw, "Unsupported", http.StatusNotImplemented)
	}
	m.NextHandler.After(rw, r, ctx)
}
