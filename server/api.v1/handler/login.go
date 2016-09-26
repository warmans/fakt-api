package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/warmans/ctxhandler"
	"github.com/warmans/stressfaktor-api/server/api.v1/common"
	"github.com/warmans/stressfaktor-api/server/data/store"
	"golang.org/x/net/context"
)

func NewLoginHandler(users *store.UserStore, sess sessions.Store) ctxhandler.CtxHandler {
	return &LoginHandler{users: users, sessions: sess}
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginHandler struct {
	users    *store.UserStore
	sessions sessions.Store
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	payload := &LoginPayload{}
	json.NewDecoder(r.Body).Decode(payload)

	if payload.Username == "" || payload.Password == "" {
		common.SendError(rw, common.HTTPError{"Username or password missing", http.StatusBadRequest, nil}, false)
		return
	}

	user, err := h.users.Authenticate(payload.Username, payload.Password)
	if err != nil {
		common.SendError(rw, common.HTTPError{"Authentication failed due to an internal error", http.StatusInternalServerError, err}, true)
		return
	}
	if user == nil {
		common.SendError(rw, common.HTTPError{"Unknown user", http.StatusOK, err}, false)
		return
	}

	//user is authenticated but double check they have an ID in case some return is missed above
	if user.ID < 1 {
		common.SendError(rw, common.HTTPError{"Unkonwn error", http.StatusOK, fmt.Errorf("User had a zero ID. This should not happen: %+v", user)}, true)
		return
	}

	//create their session
	sess, err := h.sessions.Get(r, "login")
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to create session", http.StatusInternalServerError, err}, true)
		return
	}
	sess.Values["userId"] = user.ID
	sess.Save(r, rw)

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: user,
		},
	)
}