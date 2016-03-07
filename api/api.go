package api

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/handler"
	"github.com/warmans/stressfaktor-api/data/store"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

type API struct {
	AppVersion   string
	EventStore   *store.Store
	AuthStore    *store.AuthStore
	SessionStore sessions.Store
}

func (a *API) NewServeMux() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/event", handler.NewEventHandler(a.EventStore))
	mux.Handle("/event_type", handler.NewEventTypeHandler(a.EventStore))
	mux.Handle("/venue", handler.NewVenueHandler(a.EventStore))
	mux.Handle("/performer", handler.NewPerformerHandler(a.EventStore))

	//user
	mux.Handle("/login", handler.NewLoginHandler(a.AuthStore, a.SessionStore))
	mux.Handle("/register", handler.NewRegisterHandler(a.AuthStore, a.SessionStore))
	mux.Handle("/me", handler.NewMeHandler(a.SessionStore))

	//meta
	mux.Handle("/version", handler.NewVersionHandler(a.AppVersion))

	return context.ClearHandler(mux)
}

