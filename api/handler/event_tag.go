package handler

import (
	"github.com/warmans/stressfaktor-api/api/common"
	"net/http"
	"golang.org/x/net/context"
	"github.com/gorilla/mux"
	"github.com/warmans/stressfaktor-api/data/store"
)

func NewEventTagHandler(eventStore *store.Store) common.CtxHandler {
	return &EventTagHandler{EventStore: eventStore}
}

type EventTagHandler struct{
	EventStore *store.Store
}

func (h *EventTagHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	defer r.Body.Close()

	if r.Method != "POST" {
		common.SendError(rw, common.HTTPError{"Not implemented", http.StatusMethodNotAllowed, nil}, false)
		return
	}

	vars := mux.Vars(r)
	eventId := vars["id"]

	user := ctx.Value("user");
	if user == nil {
		common.SendError(rw, common.HTTPError{"Not logged in", http.StatusForbidden, nil}, false)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status: http.StatusOK,
			Payload: struct{},
		},
	)
}
