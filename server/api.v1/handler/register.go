package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"golang.org/x/net/context"
	"github.com/warmans/fakt-api/server/data/service/user"
	"github.com/go-kit/kit/log"
)

func NewRegisterHandler(users *user.UserStore, sess sessions.Store, logger log.Logger) ctxhandler.CtxHandler {
	return &RegisterHandler{users: users, sessions: sess}
}

type RegisterPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterHandler struct {
	users    *user.UserStore
	sessions sessions.Store
}

func (h *RegisterHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	logger, ok := ctx.Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	payload := &RegisterPayload{}
	json.NewDecoder(r.Body).Decode(payload)

	if payload.Username == "" || payload.Password == "" {
		common.SendError(rw, common.HTTPError{"Username or password missing", http.StatusBadRequest, nil}, nil)
		return
	}

	user, err := h.users.Register(payload.Username, payload.Password)
	if err != nil {
		common.SendError(rw, common.HTTPError{"Registration failed due to an internal error", http.StatusInternalServerError, err}, logger)
		return
	}
	if user == nil {
		common.SendError(rw, common.HTTPError{"Unknown error", http.StatusInternalServerError, nil}, nil)
		return
	}

	//user is authenticated but double check they have an ID in case some return is missed above
	if user.ID < 1 {
		common.SendError(rw, common.HTTPError{"Unkonwn error", http.StatusOK, fmt.Errorf("User had a zero ID. This should not happen: %+v", user)}, logger)
		return
	}

	//create their session
	sess, err := h.sessions.Get(r, "login")
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to create session", http.StatusInternalServerError, err}, logger)
		return
	}

	sess.Values["user"] = user.ID
	sess.Save(r, rw)

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: user,
			Message: "Registration Successful",
		},
	)
}
