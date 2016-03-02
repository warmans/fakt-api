package api

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/handler"
	"github.com/warmans/stressfaktor-api/data/store"
)

type API struct {
	AppVersion    string
	EventStore *store.Store
}

func (a *API) NewServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/event", handler.NewEventHandler(a.EventStore))
	mux.Handle("/event_type", handler.NewEventTypeHandler(a.EventStore))
	mux.Handle("/venue", handler.NewVenueHandler(a.EventStore))
	mux.Handle("/performer", handler.NewPerformerHandler(a.EventStore))
	mux.Handle("/version", handler.NewVersionHandler(a.AppVersion))
	return mux
}
