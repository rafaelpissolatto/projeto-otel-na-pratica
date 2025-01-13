// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// PlanHandler is an HTTP handler that performs CRUD operations for model.Plan using a store.Plan
type PlanHandler struct {
	store  store.Plan
	tracer trace.Tracer
}

// NewPlanHandler returns a new PlanHandler
func NewPlanHandler(store store.Plan) *PlanHandler {
	tracer := otel.Tracer("plan handler")
	return &PlanHandler{
		store:  store,
		tracer: tracer,
	}
}

func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "PlanHandler.List")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("List plans")

	plans, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(plans)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "PlanHandler.Create")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Create plans")

	plan := &model.Plan{}
	if err := json.NewDecoder(r.Body).Decode(plan); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(r.Context(), plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(created)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PlanHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "PlanHandler.Get")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Get plans")

	id := r.PathValue("id")
	plan, err := h.store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "PlanHandler.Update")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Update plans")

	plan := &model.Plan{}
	if err := json.NewDecoder(r.Body).Decode(plan); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updated, err := h.store.Update(r.Context(), plan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(updated)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PlanHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "PlanHandler.Delete")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Delete plans")

	id := r.PathValue("id")
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
