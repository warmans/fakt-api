package handler
import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
	"github.com/warmans/stressfaktor-api/data/store"
	"github.com/gorilla/sessions"
	"fmt"
"golang.org/x/net/context"
)

func NewRegisterHandler(auth *store.AuthStore, sess sessions.Store) common.CtxHandler {
	return &RegisterHandler{auth: auth, sessions: sess}
}

type RegisterHandler struct {
	auth     *store.AuthStore
	sessions sessions.Store
}

func (h *RegisterHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	defer r.Body.Close()
	r.ParseForm()

	if r.Method != "POST" {
		common.SendError(rw, common.HTTPError{"Only POST requests are supported for login", http.StatusMethodNotAllowed, nil}, false)
		return
	}

	if r.Form.Get("username") == "" || r.Form.Get("password") == "" {
		common.SendError(rw, common.HTTPError{"Username or password missing", http.StatusBadRequest, nil}, false)
		return
	}

	user, err := h.auth.Register(r.Form.Get("username"), r.Form.Get("password"))
	if err != nil {
		common.SendError(rw, common.HTTPError{"Registration failed due to an internal error", http.StatusInternalServerError, err}, true)
		return
	}
	if user == nil {
		common.SendError(rw, common.HTTPError{"Unknown error", http.StatusInternalServerError, nil}, false)
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

	sess.Values["user"] = user.ID
	sess.Save(r, rw)

	common.SendResponse(
		rw,
		&common.Response{
			Status: http.StatusOK,
			Payload: user,
		},
	)
}
