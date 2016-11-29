package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/sessions"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/service/user"
)

func NewLoginHandler(users *user.UserStore, sess sessions.Store, logger log.Logger) http.Handler {
	return &LoginHandler{users: users, sessions: sess, logger: logger}
}

type LoginPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginHandler struct {
	users    *user.UserStore
	sessions sessions.Store
	logger   log.Logger
}

func (h *LoginHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	payload := &LoginPayload{}
	json.NewDecoder(r.Body).Decode(payload)

	if payload.Username == "" || payload.Password == "" {
		common.SendError(rw, common.HTTPError{"Username or password missing", http.StatusBadRequest, nil}, nil)
		return
	}

	usr, err := h.users.Authenticate(payload.Username, payload.Password)
	if err != nil {
		common.SendError(rw, common.HTTPError{"Authentication failed due to an internal error", http.StatusInternalServerError, err}, h.logger)
		return
	}
	if usr == nil {
		common.SendError(rw, common.HTTPError{"Unknown user", http.StatusOK, err}, nil)
		return
	}

	//user is authenticated but double check they have an ID in case some return is missed above
	if usr.ID < 1 {
		common.SendError(rw, common.HTTPError{"Unkonwn error", http.StatusOK, fmt.Errorf("User had a zero ID. This should not happen: %+v", usr)}, h.logger)
		return
	}

	//create their session
	sess, err := h.sessions.Get(r, "login")
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to create session", http.StatusInternalServerError, err}, h.logger)
		return
	}
	sess.Values["userId"] = usr.ID

	if err := sess.Save(r, rw); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save session", http.StatusInternalServerError, err}, h.logger)
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
