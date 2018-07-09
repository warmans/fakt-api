package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"go.uber.org/zap"
)

type commonContextKey string

func AddCommonCtx(nextHandler http.Handler, sess sessions.Store, logger *zap.Logger) http.Handler {
	return &CommonCtxMiddleware{next: nextHandler, sessions: sess, logger: logger}
}

type CommonCtxMiddleware struct {
	next     http.Handler
	sessions sessions.Store
	logger   *zap.Logger
}

func (m *CommonCtxMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	//setup logger with http context info and add to context
	logger := m.logger.With(zap.String("method", r.Method), zap.String("url", r.URL.String()))
	ctx = context.WithValue(ctx, commonContextKey("logger"), logger)

	m.next.ServeHTTP(rw, r.WithContext(ctx))
}

func MustGetLogger(r *http.Request) *zap.Logger {
	logger, ok := r.Context().Value(commonContextKey("logger")).(*zap.Logger)
	if !ok {
		panic("Context must contain a logger")
	}
	return logger
}
