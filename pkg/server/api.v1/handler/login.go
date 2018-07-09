package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/warmans/fakt-api/pkg/server/api.v1/common"
	"github.com/warmans/fakt-api/pkg/server/data/service/user"
	"github.com/warmans/fakt-api/pkg/server/api.v1/middleware"
)

func NewLoginHandler(users *user.UserStore, sess sessions.Store) http.Handler {
	return &LoginHandler{users: users, sessions: sess}
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginHandler struct {
	users    *user.UserStore
	sessions sessions.Store
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	payload := &LoginPayload{}
	json.NewDecoder(r.Body).Decode(payload)

	if payload.Username == "" || payload.Password == "" {
		common.SendError(rw, common.HTTPError{"Username or password missing", http.StatusBadRequest, nil}, nil)
		return
	}

	usr, err := h.users.Authenticate(payload.Username, payload.Password)
	if err != nil {
		common.SendError(rw, common.HTTPError{"Authentication failed due to an internal error", http.StatusInternalServerError, err}, logger)
		return
	}
	if usr == nil {
		common.SendError(rw, common.HTTPError{"Unknown user", http.StatusOK, err}, nil)
		return
	}

	//user is authenticated but double check they have an ID in case some return is missed above
	if usr.ID < 1 {
		common.SendError(rw, common.HTTPError{"Unkonwn error", http.StatusOK, fmt.Errorf("User had a zero ID. This should not happen: %+v", usr)}, logger)
		return
	}

	//create their session
	sess, err := h.sessions.Get(r, "login")
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to create session", http.StatusInternalServerError, err}, logger)
		return
	}
	sess.Values["userId"] = usr.ID

	if err := sess.Save(r, rw); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save session", http.StatusInternalServerError, err}, logger)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: usr,
		},
	)
}
