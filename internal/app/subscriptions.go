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
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm/logger"
)

type Subscription struct {
	Handler *subscriptionhttp.SubscriptionHandler
	Store   store.Subscription
	Tracer  trace.Tracer
	Meter   metric.Meter
	Logger  *logrus.Logger
}

func NewSubscription(cfg *config.Subscriptions) *Subscription {
	store := memory.NewSubscriptionStore()
	tracer := otel.Tracer("subscription")
	meter := otel.Meter("subscription")
	logger := telemetry.GetLogger()

	return &Subscription{
		Handler: subscriptionhttp.NewSubscriptionHandler(store, cfg.UsersEndpoint, cfg.PlansEndpoint),
		Store:   store,
		Tracer:  tracer,
		Meter:   meter,
		Logger:  logger,
	}
}

func (a *Subscription) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /subscriptions", a.traceHandler(a.metricHandler(a.logHandler(a.Handler.List))))
	mux.HandleFunc("POST /subscriptions", a.traceHandler(a.metricHandler(a.Handler.Create)))
	mux.HandleFunc("GET /subscriptions/{id}", a.traceHandler(a.metricHandler(a.Handler.Get)))
	mux.HandleFunc("PUT /subscriptions/{id}", a.traceHandler(a.metricHandler(a.Handler.Update)))
	mux.HandleFunc("DELETE /subscriptions/{id}", a.traceHandler(a.metricHandler(a.Handler.Delete)))
}

func (a *Subscription) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}

func (a *Subscription) metricHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counter, err := a.Meter.Int64Counter("subscription.requests")
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
		logger.Default.Info(r.Context(), "request", r.Method+" "+r.URL.Path)
		handlerFunc(w, r)
	}
}
