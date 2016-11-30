package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/api.v1/common"
	dcom "github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"github.com/warmans/route-rest/routes"
)

func NewPerformerTagHandler(ds *utag.UTagService, logger log.Logger) routes.RESTHandler {
	return &PerformerTagHandler{ds: ds, logger: logger}
}

type PerformerTagHandler struct {
	routes.DefaultRESTHandler
	ds     *utag.UTagService
	logger log.Logger
}

func (h *PerformerTagHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
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

func (h *PerformerTagHandler) HandlePost(rw http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	usr, err := common.Restrict(r.Context())
	if err != nil {
		common.SendError(rw, err, nil)
		return
	}

	payload := make([]string, 0)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, nil)
		return
	}

	if err := h.ds.StorePerformerUTags(int64(performerID), usr.ID, payload); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, h.logger)
		return
	}

	h.HandleGetList(rw, r)
}

func (h *PerformerTagHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	performerID, err := strconv.Atoi(vars["performer_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid performerID", http.StatusBadRequest, err}, nil)
		return
	}

	usr, err := common.Restrict(r.Context())
	if err != nil {
		common.SendError(rw, err, nil)
		return
	}

	payload := make([]string, 0)
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, nil)
		return
	}

	if err := h.ds.RemovePerformerUTags(int64(performerID), usr.ID, payload); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, h.logger)
		return
	}

	h.HandleGetList(rw, r)
}
