package handler

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
	"github.com/gorilla/sessions"
	"errors"
	"github.com/warmans/stressfaktor-api/data/store"
)

func NewMeHandler(sess sessions.Store) http.Handler {
	return &MeHandler{sessions: sess}
}

type MeHandler struct {
	sessions sessions.Store
	auth     *store.AuthStore
}

func (h *MeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//create their session
	sess, err := h.sessions.Get(r, "login")
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to get session", http.StatusForbidden, err}, false)
		return
	}

	userId, found := sess.Values["user"]
	if found == false {
		common.SendError(rw, common.HTTPError{"Failed to get session", http.StatusForbidden, errors.New("No user in session. This shouldn't happen.")}, true)
		return
	}

	user, err := h.auth.GetUser(userId.(int64))
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to get user", http.StatusInternalServerError, err}, true)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status: http.StatusOK,
			Payload: user,
		},
	)
}
