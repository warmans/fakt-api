package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	dcom "github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/user"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"golang.org/x/net/context"
)

func NewPerformerTagHandler(ds *utag.UTagService, logger log.Logger) ctxhandler.CtxHandler {
	return &PerformerTagHandler{ds: ds, logger: logger}
}

type PerformerTagHandler struct {
	ds     *utag.UTagService
	logger log.Logger
}

func (h *PerformerTagHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
	}

	user := ctx.Value("user").(*user.User)
	if user == nil {
		common.SendError(rw, common.HTTPError{"Not logged in", http.StatusForbidden, nil}, nil)
		return
	}

	payload := make([]string, 0)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, nil)
		return
	}

	//store any submitted tags
	if r.Method == "POST" {
		if err := h.ds.StorePerformerUTags(int64(performerID), user.ID, payload); err != nil {
			common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, h.logger)
			return
		}
	}
	if r.Method == "DELETE" {
		if err := h.ds.RemovePerformerUTags(int64(performerID), user.ID, payload); err != nil {
			common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, h.logger)
			return
		}
	}

	//then get all tags for the event
	tags, err := h.ds.FindPerformerUTags(int64(performerID), &dcom.UTagsFilter{Username: r.Form.Get("username")})
	if err != nil && err != sql.ErrNoRows {
		common.SendError(rw, common.HTTPError{"Failed to get tags", http.StatusInternalServerError, err}, h.logger)
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
