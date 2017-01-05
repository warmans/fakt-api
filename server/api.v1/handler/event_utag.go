package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/warmans/fakt-api/server/api.v1/common"
	dcom "github.com/warmans/fakt-api/server/data/service/common"
	"github.com/warmans/fakt-api/server/data/service/utag"
	"github.com/warmans/route-rest/routes"
	"github.com/warmans/fakt-api/server/api.v1/middleware"
)

func NewEventUTagHandler(uts *utag.UTagService) routes.RESTHandler {
	return &EventUTagHandler{uts: uts}
}

type EventUTagHandler struct {
	uts    *utag.UTagService
	routes.DefaultRESTHandler
}

func (h *EventUTagHandler) HandleGetList(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["event_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, nil)
		return
	}

	//then get all tags for the event
	tags, err := h.uts.FindEventUTags(int64(eventId), &dcom.UTagsFilter{Username: r.Form.Get("username")})
	if err != nil && err != sql.ErrNoRows {
		common.SendError(rw, common.HTTPError{"Failed to get tags", http.StatusInternalServerError, err}, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: tags})
}

func (h *EventUTagHandler) HandlePost(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["event_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, nil)
		return
	}

	usr, err := middleware.Restrict(r)
	if err != nil {
		common.SendError(rw, err, nil)
		return
	}

	payload := make([]string, 0)
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, nil)
		return
	}

	if err := h.uts.StoreEventUTags(int64(eventId), usr.ID, payload); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, logger)
		return
	}

	h.HandleGetList(rw, r)
}

func (h *EventUTagHandler) HandleDelete(rw http.ResponseWriter, r *http.Request) {

	logger := middleware.MustGetLogger(r)

	vars := mux.Vars(r)
	eventId, err := strconv.Atoi(vars["event_id"])
	if err != nil {
		common.SendError(rw, common.HTTPError{"Invalid eventID", http.StatusBadRequest, err}, nil)
		return
	}

	usr, err := middleware.Restrict(r)
	if err != nil {
		common.SendError(rw, err, nil)
		return
	}

	payload := make([]string, 0)
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		common.SendError(rw, common.HTTPError{"Invalid payload", http.StatusBadRequest, nil}, nil)
		return
	}

	if err := h.uts.RemoveEventUTags(int64(eventId), usr.ID, payload); err != nil {
		common.SendError(rw, common.HTTPError{"Failed to save tags", http.StatusInternalServerError, err}, logger)
		return
	}

	h.HandleGetList(rw, r)
}
