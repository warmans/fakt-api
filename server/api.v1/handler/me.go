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

	user := ctx.Value("user")
	if user == nil {
		common.SendError(rw, common.HTTPError{"Not logged in", http.StatusForbidden, nil}, false)
		return
	}

	common.SendResponse(
		rw,
		&common.Response{
			Status:  http.StatusOK,
			Payload: user,
		},
	)
}
