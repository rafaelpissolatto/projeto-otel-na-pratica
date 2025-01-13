// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	userhttp "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/handler/http"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store/memory"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type User struct {
	Handler   *userhttp.UserHandler
	Store     store.User
	Telemetry *telemetry.Telemetry
}

func NewUser(ctx context.Context, _ *config.Users, telemetry *telemetry.Telemetry) *User {
	ctx, span := otel.Tracer("user").Start(ctx, "NewUser")
	defer span.End()

	store := memory.NewUserStore(ctx)
	return &User{
		Handler:   userhttp.NewUserHandler(store, telemetry),
		Store:     store,
		Telemetry: telemetry,
	}
}

func (a *User) RegisterRoutes(mux *http.ServeMux) {
	// Register the routes for the user service
	handleFunc(mux, "GET /users", a.chainHandlers(a.Handler.List))
	handleFunc(mux, "POST /users", a.chainHandlers(a.Handler.Create))
	handleFunc(mux, "GET /users/{id}", a.chainHandlers(a.Handler.Get))
	handleFunc(mux, "PUT /users/{id}", a.chainHandlers(a.Handler.Update))
	handleFunc(mux, "DELETE /users/{id}", a.chainHandlers(a.Handler.Delete))
}

// handleFunc is a replacement for mux.HandleFunc
// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
// This is useful for the observability of the application.
func handleFunc(mux *http.ServeMux, pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	// Configure the "http.route" for the HTTP instrumentation.
	handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
	mux.Handle(pattern, handler)
}

func (a *User) chainHandlers(handlerFun http.HandlerFunc) http.HandlerFunc {
	return a.traceHandler(a.metricHandler(a.logHandler(handlerFun)))
}

func (a *User) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Telemetry.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}

func (a *User) metricHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counter, err := a.Telemetry.Meter.Int64Counter("user.requests")
		if err != nil {
			http.Error(w, "failed to create counter", http.StatusInternalServerError)
			return
		}
		counter.Add(r.Context(), 1)
		handlerFunc(w, r)
	}
}

func (a *User) logHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// logger.Info(r.Context(), "request", r.Method+" "+r.URL.Path)
		a.Telemetry.Logger.Info("request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
		handlerFunc(w, r)
	}
}
