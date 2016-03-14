package handler

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
	"github.com/warmans/stressfaktor-api/data/store"
	"github.com/gorilla/sessions"
	"fmt"
	"golang.org/x/net/context"
)

func NewLogoutHandler(sess sessions.Store) common.CtxHandler {
	return &LogoutHandler{sessions: sess}
}

type LogoutHandler struct {
	auth     *store.AuthStore
	sessions sessions.Store
}

func (h *LogoutHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	defer r.Body.Close()
	r.ParseForm()

	//create their session
	sess, err := h.sessions.Get(r, "login")
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to create session", http.StatusInternalServerError, err}, true)
		return
	}
	delete(sess.Values, "user")
	sess.Save(r, rw)

	common.SendResponse(
		rw,
		&common.Response{
			Status: http.StatusOK,
			Payload: nil,
		},
	)
}
