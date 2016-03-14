package handler

import (
	"github.com/warmans/stressfaktor-api/api/common"
	"net/http"
	"golang.org/x/net/context"
	"github.com/gorilla/mux"
	"github.com/warmans/stressfaktor-api/data/store"
	"database/sql"
	"strconv"
	"strings"
)

func NewEventTagHandler(eventStore *store.Store) common.CtxHandler {
	return &EventTagHandler{EventStore: eventStore}
}

type EventTagHandler struct{
	EventStore *store.Store
}

func (h *EventTagHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	defer r.Body.Close()
	r.ParseForm()

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, false)
	}

	user := ctx.Value("user").(*store.User);
	if user == nil {
		common.SendError(rw, common.HTTPError{"Not logged in", http.StatusForbidden, nil}, false)
		return
	}

	//store any submitted tags
	if r.Method == "POST" {
		if err := h.EventStore.StoreEventUTags(int64(eventId), user.ID, strings.Split(r.Form.Get("tags"), ";")); err != nil {
			common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, true)
			return
		}
	}

	//then get all tags for the event
	tags, err := h.EventStore.FindEventUTags(int64(eventId))
	if err != nil && err != sql.ErrNoRows {
		common.SendError(rw, common.HTTPError{"Failed to get tags", http.StatusInternalServerError, err}, true)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status: http.StatusOK,
			Payload: tags,
		},
	)
}
