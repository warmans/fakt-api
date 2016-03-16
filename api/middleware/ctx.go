package middleware
import (
	"github.com/gorilla/sessions"
	"github.com/warmans/stressfaktor-api/data/store"
	"net/http"
"golang.org/x/net/context"
	"log"
	"github.com/warmans/stressfaktor-api/api/common"
)

func AddCtx(handler common.CtxHandler, sess sessions.Store, auth *store.AuthStore, restrict bool) http.Handler {
	return &CtxMiddleware{NextHandler: handler, SessionStore: sess, AuthStore: auth}
}

type CtxMiddleware struct {
	NextHandler    common.CtxHandler
	SessionStore   sessions.Store
	AuthStore      *store.AuthStore
	RestrictAccess bool
}

func (m *CtxMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	sess, err := m.SessionStore.Get(r, "login")
	if err != nil {
		log.Printf("Failed to get session: %s", err.Error())
		m.NextHandler.ServeHTTP(rw, r, ctx)
		return
	}

	userId, found := sess.Values["userId"]
	if found == false || userId == nil || userId.(int64) < 1 {
		if m.RestrictAccess {
			common.SendError(rw,common.HTTPError{"Access Denied", http.StatusForbidden, nil}, false)
			return
		}
		m.NextHandler.ServeHTTP(rw, r, ctx)
		return
	}

	user, err := m.AuthStore.GetUser(userId.(int64))
	if err == nil && user != nil {
		ctx = context.WithValue(ctx, "user", user)
	} else {
		if m.RestrictAccess {
			common.SendError(rw, common.HTTPError{"Access Denied", http.StatusForbidden, nil}, false)
			return
		}
	}

	m.NextHandler.ServeHTTP(rw, r, ctx)
}
