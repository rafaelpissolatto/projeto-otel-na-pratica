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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Plan struct {
	Handler     *planhttp.PlanHandler
	GRPCHandler api.PlanServiceServer
	Store       store.Plan
	Tracer      trace.Tracer
}

func NewPlan(*config.Plans) *Plan {
	store := memory.NewPlanStore()
	tracer := otel.Tracer("plan")
	return &Plan{
		Handler:     planhttp.NewPlanHandler(store),
		GRPCHandler: grpchandler.NewPlanServer(store),
		Store:       store,
		Tracer:      tracer,
	}
}

func (a *Plan) RegisterRoutes(mux *http.ServeMux, grpcSrv *grpc.Server) {
	mux.HandleFunc("GET /plans", a.traceHandler(a.Handler.List))
	mux.HandleFunc("POST /plans", a.traceHandler(a.Handler.Create))
	mux.HandleFunc("GET /plans/{id}", a.traceHandler(a.Handler.Get))
	mux.HandleFunc("PUT /plans/{id}", a.traceHandler(a.Handler.Update))
	mux.HandleFunc("DELETE /plans/{id}", a.traceHandler(a.Handler.Delete))

	api.RegisterPlanServiceServer(grpcSrv, a.GRPCHandler)
}

func (a *Plan) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}
