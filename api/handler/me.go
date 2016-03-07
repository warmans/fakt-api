package handler
import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/common"
	"github.com/gorilla/sessions"
	"errors"
	"log"
)

func NewMeHandler(sess sessions.Store) http.Handler {
	return &MeHandler{sessions: sess}
}

type MeHandler struct {
	sessions sessions.Store
}

func (h *MeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	//create their session
	sess, err := h.sessions.Get(r, "login")
	if err != nil {
		common.SendError(rw, common.HTTPError{"Failed to get session", http.StatusForbidden, err}, false)
		return
	}
	sess.Save(r, rw)
	log.Printf("%+v", sess)

	user, found := sess.Values["user"]
	if found == false {
		common.SendError(rw, common.HTTPError{"Failed to get session", http.StatusForbidden, errors.New("No user in session. This shouldn't happen.")}, true)
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
