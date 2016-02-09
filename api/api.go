package api

import (
	"github.com/warmans/stressfaktor-api/entity"
	"net/http"
	"github.com/warmans/stressfaktor-api/api/handler"
)

type API struct {
	AppVersion    string
	EventStore *entity.EventStore
}

func (a *API) NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/event", handler.NewEventHandler(a.EventStore))
	mux.Handle("/version", handler.NewVersionHandler(a.AppVersion))
	return mux
}
