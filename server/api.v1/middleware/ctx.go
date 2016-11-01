package middleware

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/sessions"
	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/user"
	"golang.org/x/net/context"
)

func AddCtx(nextHandler ctxhandler.CtxHandler, sess sessions.Store, users *user.UserStore, restrict bool, logger log.Logger) http.Handler {
	return &CtxMiddleware{next: nextHandler, sessions: sess, users: users, restrict: restrict, logger: logger}
}

type CtxMiddleware struct {
	next     ctxhandler.CtxHandler
	sessions sessions.Store
	users    *user.UserStore
	restrict bool
	logger   log.Logger
}

func (m *CtxMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	//setup logger with http context info and add to context
	logger := log.NewContext(m.logger).With("method", r.Method, "url", r.URL.String())
	ctx = context.WithValue(ctx, "logger", m.logger)

	sess, err := m.sessions.Get(r, "login")
	if err != nil {
		logger.Log("Failed to get session: %s", err.Error())
		m.next.ServeHTTP(rw, r, ctx)
		return
	}

	userId, found := sess.Values["userId"]
	if found == false || userId == nil || userId.(int64) < 1 {
		if m.restrict {
			common.SendError(rw, common.HTTPError{"Access Denied", http.StatusUnauthorized, nil}, nil)
			return
		}
		m.next.ServeHTTP(rw, r, ctx)
		return
	}

	user, err := m.users.GetUser(userId.(int64))
	if err == nil && user != nil {
		ctx = context.WithValue(ctx, "user", user)
	} else {
		if m.restrict {
			common.SendError(rw, common.HTTPError{"Access Denied", http.StatusUnauthorized, nil}, nil)
			return
		}
	}

	m.next.ServeHTTP(rw, r, ctx)
}
