package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"github.com/warmans/fakt-api/server/data/store"
	"golang.org/x/net/context"
)

func NewPerformerHandler(ds *store.Store) ctxhandler.CtxHandler {
	return &PerformerHandler{ds: ds}
}

type PerformerHandler struct {
	ds *store.Store
}

func (h *PerformerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	//query to filter
	filter := &store.PerformerFilter{PerformerID: make([]int, 0)}
	if venue := r.Form.Get("performer"); venue != "" {
		for _, idStr := range strings.Split(venue, ",") {
			if idInt, err := strconv.Atoi(idStr); err == nil {
				filter.PerformerID = append(filter.PerformerID, idInt)
			}
		}
	}
	filter.Name = r.Form.Get("name")
	filter.Genre = r.Form.Get("genre")
	filter.Home = r.Form.Get("home")

	performers, err := h.ds.FindPerformers(filter)
	if err != nil {
		log.Print(err.Error())
		common.SendResponse(rw, &common.Response{Status: http.StatusInternalServerError, Payload: nil})
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: performers})
}
