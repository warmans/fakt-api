package handler

import (
	"net/http"

	"github.com/warmans/ctxhandler"
	"github.com/warmans/fakt-api/server/api.v1/common"
	"golang.org/x/net/context"
)

func NewMeHandler() ctxhandler.CtxHandler {
	return &MeHandler{}
}

type MeHandler struct{}

func (h *MeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request, ctx context.Context) {
	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: ctx.Value("user"),
		},
	)
}
