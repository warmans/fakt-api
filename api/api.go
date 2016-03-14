package api

import (
	"net/http"
	"github.com/warmans/stressfaktor-api/api/handler"
	"github.com/warmans/stressfaktor-api/data/store"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
"github.com/warmans/stressfaktor-api/api/common"
	"github.com/gorilla/mux"
)

type API struct {
	AppVersion   string
	EventStore   *store.Store
	AuthStore    *store.AuthStore
	SessionStore sessions.Store
}

func (a *API) NewServeMux() http.Handler {
	mux := mux.NewRouter()

	mux.Handle(
		"/event",
		common.AddCtx(handler.NewEventHandler(a.EventStore), a.SessionStore, a.AuthStore),
	)
	mux.Handle(
		"/event/{id:[0-9]+}/tag",
		common.AddCtx(handler.NewEventTagHandler(a.EventStore), a.SessionStore, a.AuthStore),
	)
	mux.Handle(
		"/event_type",
		common.AddCtx(handler.NewEventTypeHandler(a.EventStore), a.SessionStore, a.AuthStore),
	)
	mux.Handle(
		"/venue",
		common.AddCtx(handler.NewVenueHandler(a.EventStore), a.SessionStore, a.AuthStore),
	)
	mux.Handle(
		"/performer",
		common.AddCtx(handler.NewPerformerHandler(a.EventStore), a.SessionStore, a.AuthStore),
	)

	//user
	mux.Handle(
		"/login",
		common.AddCtx(handler.NewLoginHandler(a.AuthStore, a.SessionStore), a.SessionStore, a.AuthStore),
	)
	mux.Handle(
		"/logout",
		common.AddCtx(handler.NewLogoutHandler(a.SessionStore), a.SessionStore, a.AuthStore),
	)

	mux.Handle(
		"/register",
		common.AddCtx(handler.NewRegisterHandler(a.AuthStore, a.SessionStore), a.SessionStore, a.AuthStore),
	)
	mux.Handle(
		"/me",
		common.AddCtx(handler.NewMeHandler(), a.SessionStore, a.AuthStore),
	)

	//meta
	mux.Handle("/version", handler.NewVersionHandler(a.AppVersion))

	return context.ClearHandler(mux)
}

