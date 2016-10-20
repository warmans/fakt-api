package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"golang.org/x/net/context"
	"github.com/warmans/fakt-api/server/data/service/performer"
	"github.com/go-kit/kit/log"
)

func NewPerformerHandler(ds *performer.PerformerService) ctxhandler.CtxHandler {
	return &PerformerHandler{ds: ds}
}

type PerformerHandler struct {
	ds *performer.PerformerService
}

func (h *PerformerHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {

	logger, ok := ctx.Value("logger").(log.Logger)
	if !ok {
		panic("Context must contain a logger")
	}

	//query to filter
	filter := &performer.PerformerFilter{PerformerID: make([]int, 0)}
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
		common.SendError(rw, err, logger)
		return
	}

	common.SendResponse(rw, &common.Response{Status: http.StatusOK, Payload: performers})
}
