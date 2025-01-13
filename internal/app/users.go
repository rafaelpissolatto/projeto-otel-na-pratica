// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	userhttp "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/handler/http"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store/memory"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type User struct {
	Handler *userhttp.UserHandler
	Store   store.User
	Tracer  trace.Tracer
}

func NewUser(*config.Users) *User {
	store := memory.NewUserStore()
	tracer := otel.Tracer("user")
	return &User{
		Handler: userhttp.NewUserHandler(store),
		Store:   store,
		Tracer:  tracer,
	}
}

func (a *User) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", a.traceHandler(a.Handler.List))
	mux.HandleFunc("POST /users", a.traceHandler(a.Handler.Create))
	mux.HandleFunc("GET /users/{id}", a.traceHandler(a.Handler.Get))
	mux.HandleFunc("PUT /users/{id}", a.traceHandler(a.Handler.Update))
	mux.HandleFunc("DELETE /users/{id}", a.traceHandler(a.Handler.Delete))
}

func (a *User) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}
