package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/store"
	"golang.org/x/net/context"
)

func NewEventTagHandler(ds *store.Store) ctxhandler.CtxHandler {
	return &EventTagHandler{ds: ds}
}

type EventTagHandler struct {
	ds *store.Store
}

func (h *EventTagHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, false)
	}

	user := ctx.Value("user").(*store.User)
	if user == nil {
		common.SendError(rw, common.HTTPError{"Not logged in", http.StatusForbidden, nil}, false)
		return
	}

	payload := make([]string, 0)
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, false)
		return
	}

	//store any submitted tags
	if r.Method == "POST" {
		if err := h.ds.StoreEventUTags(int64(eventId), user.ID, payload); err != nil {
			common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, true)
			return
		}
	}
	if r.Method == "DELETE" {
		if err := h.ds.RemoveEventUTags(int64(eventId), user.ID, payload); err != nil {
			common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, true)
			return
		}
	}

	//then get all tags for the event
	tags, err := h.ds.FindEventUTags(int64(eventId), &store.UTagsFilter{Username: r.Form.Get("username")})
	if err != nil && err != sql.ErrNoRows {
		common.SendError(rw, common.HTTPError{"Failed to get tags", http.StatusInternalServerError, err}, true)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: tags,
		},
	)
}
