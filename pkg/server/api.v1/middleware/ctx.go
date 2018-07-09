package middleware

import (
	"net/http"

	"context"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/sessions"
	"github.com/warmans/fakt-api/pkg/server/data/service/user"
	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
)

type commonContextKey string

func AddCommonCtx(nextHandler http.Handler, sess sessions.Store, users *user.UserStore, logger log.Logger) http.Handler {
	return &CommonCtxMiddleware{next: nextHandler, sessions: sess, users: users, logger: logger}
}

type CommonCtxMiddleware struct {
	next     http.Handler
	sessions sessions.Store
	users    *user.UserStore
	logger   log.Logger
}

func (m *CommonCtxMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	//setup logger with http context info and add to context
	logger := log.NewContext(m.logger).With("method", r.Method, "url", r.URL.String())
	ctx = context.WithValue(ctx, commonContextKey("logger"), m.logger)

	sess, err := m.sessions.Get(r, "login")
	if err != nil {
		logger.Log("msg", "Failed to get session: "+err.Error())
		m.next.ServeHTTP(rw, r.WithContext(ctx))
		return
	}

	userId, found := sess.Values["userId"]
	if found == false || userId == nil || userId.(int64) < 1 {
		m.next.ServeHTTP(rw, r.WithContext(ctx))
		return
	}

	usr, err := m.users.GetUser(userId.(int64))
	if err == nil && usr != nil {
		ctx = context.WithValue(ctx, commonContextKey("user"), usr)
	} else {
		ctx = context.WithValue(ctx, commonContextKey("user"), nil)
	}

	m.next.ServeHTTP(rw, r.WithContext(ctx))
}

func MustGetLogger(r *http.Request) log.Logger {
	logger, ok := r.Context().Value(commonContextKey("logger")).(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}
	return logger
}

func GetUser(r *http.Request) *user.User {
	usr, ok := r.Context().Value(commonContextKey("user")).(*user.User)
	if !ok {
		return nil
	}
	return usr
}

func Restrict(r *http.Request) (*user.User, error) {
	usr := GetUser(r)
	if usr != nil {
		return nil, common.HTTPError{Msg:"Access Denied", Status: http.StatusUnauthorized, LastErr: nil}
	}
	return usr, nil
}
