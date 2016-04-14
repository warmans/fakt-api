package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/stressfaktor-api/server/api/common"
	"github.com/warmans/stressfaktor-api/server/data/store"
	"golang.org/x/net/context"
)

func NewPerformerTagHandler(ds *store.Store) common.CtxHandler {
	return &PerformerTagHandler{ds: ds}
}

type PerformerTagHandler struct {
	ds *store.Store
}

func (h *PerformerTagHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, false)
	}

	user := ctx.Value("user").(*store.User)
	if user == nil {
		common.SendError(rw, common.HTTPError{"Not logged in", http.StatusForbidden, nil}, false)
		return
	}

	payload := make([]string, 0)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, false)
		return
	}

	//store any submitted tags
	if r.Method == "POST" {
		if err := h.ds.StorePerformerUTags(int64(performerID), user.ID, payload); err != nil {
			common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, true)
			return
		}
	}
	if r.Method == "DELETE" {
		if err := h.ds.RemovePerformerUTags(int64(performerID), user.ID, payload); err != nil {
			common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, true)
			return
		}
	}

	//then get all tags for the event
	tags, err := h.ds.FindPerformerUTags(int64(performerID), &store.UTagsFilter{Username: r.Form.Get("username")})
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
