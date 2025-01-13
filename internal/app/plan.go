// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/api"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	grpchandler "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/handler/grpc"
	planhttp "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/handler/http"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store/memory"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Plan struct {
	Handler     *planhttp.PlanHandler
	GRPCHandler api.PlanServiceServer
	Store       store.Plan
	Telemetry   *telemetry.Telemetry
}

func NewPlan(_ *config.Plans, telemetry *telemetry.Telemetry) *Plan {
	store := memory.NewPlanStore()
	return &Plan{
		Handler:     planhttp.NewPlanHandler(store, telemetry),
		GRPCHandler: grpchandler.NewPlanServer(store),
		Store:       store,
		Telemetry:   telemetry,
	}
}

func (a *Plan) RegisterRoutes(mux *http.ServeMux, grpcSrv *grpc.Server) {
	mux.HandleFunc("GET /plans", a.chainHandlers(a.Handler.List))
	mux.HandleFunc("POST /plans", a.chainHandlers(a.Handler.Create))
	mux.HandleFunc("GET /plans/{id}", a.chainHandlers(a.Handler.Get))
	mux.HandleFunc("PUT /plans/{id}", a.chainHandlers(a.Handler.Update))
	mux.HandleFunc("DELETE /plans/{id}", a.chainHandlers(a.Handler.Delete))

	api.RegisterPlanServiceServer(grpcSrv, a.GRPCHandler)
}

func (a *Plan) chainHandlers(handlerFun http.HandlerFunc) http.HandlerFunc {
	return a.traceHandler(a.metricHandler(a.logHandler(handlerFun)))
}

func (a *Plan) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Telemetry.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}

func (a *Plan) metricHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counter, err := a.Telemetry.Meter.Int64Counter("plan.requests")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		counter.Add(r.Context(), 1)
		handlerFunc(w, r)
	}
}

func (a *Plan) logHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.Telemetry.Logger.Info("request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
		handlerFunc(w, r)
	}
}
