// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	subscriptionhttp "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/handler/http"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store/memory"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"go.uber.org/zap"
)

type Subscription struct {
	Handler   *subscriptionhttp.SubscriptionHandler
	Store     store.Subscription
	Telemetry *telemetry.Telemetry
}

func NewSubscription(cfg *config.Subscriptions, telemetry *telemetry.Telemetry) *Subscription {
	store := memory.NewSubscriptionStore()
	return &Subscription{
		Handler:   subscriptionhttp.NewSubscriptionHandler(store, cfg.UsersEndpoint, cfg.PlansEndpoint, telemetry),
		Store:     store,
		Telemetry: telemetry,
	}
}

func (a *Subscription) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /subscriptions", a.chainHandlers(a.metricHandler(a.logHandler(a.Handler.List))))
	mux.HandleFunc("POST /subscriptions", a.chainHandlers(a.metricHandler(a.Handler.Create)))
	mux.HandleFunc("GET /subscriptions/{id}", a.chainHandlers(a.metricHandler(a.Handler.Get)))
	mux.HandleFunc("PUT /subscriptions/{id}", a.chainHandlers(a.metricHandler(a.Handler.Update)))
	mux.HandleFunc("DELETE /subscriptions/{id}", a.chainHandlers(a.metricHandler(a.Handler.Delete)))
}

func (a *Subscription) chainHandlers(handlerFun http.HandlerFunc) http.HandlerFunc {
	return a.traceHandler(a.metricHandler(a.logHandler(handlerFun)))
}

func (a *Subscription) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Telemetry.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}

func (a *Subscription) metricHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counter, err := a.Telemetry.Meter.Int64Counter("subscription.requests")
		if err != nil {
			http.Error(w, "failed to create counter", http.StatusInternalServerError)
			return
		}
		counter.Add(r.Context(), 1)
		handlerFunc(w, r)
	}
}

func (a *Subscription) logHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.Telemetry.Logger.Info("request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
		handlerFunc(w, r)
	}
}
