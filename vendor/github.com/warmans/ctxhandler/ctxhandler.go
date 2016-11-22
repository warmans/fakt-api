package ctxhandler

import (
	"net/http"
	"golang.org/x/net/context"
)

//CtxHandler defines a http.Handler with context
type CtxHandler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context)
}

//Ctx wrapper function to convert a http.Handler to a CtxHandler
func Ctx(next CtxHandler) http.Handler {
	return &CtxConvertMiddleware{NextHandler: next}
}

//CtxConvertMiddleware converts a http.Handler to a CtxHandler
type CtxConvertMiddleware struct {
	NextHandler CtxHandler
}

//ServeHTTP...
func (m *CtxConvertMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	m.NextHandler.ServeHTTP(rw, r, context.Background())
}
