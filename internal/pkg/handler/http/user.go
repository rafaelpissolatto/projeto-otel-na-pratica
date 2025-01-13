// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"go.opentelemetry.io/otel/codes"
)

// UserHandler is an HTTP handler that performs CRUD operations for model.User using a store.User
type UserHandler struct {
	store     store.User
	telemetry *telemetry.Telemetry
}

// NewUserHandler returns a new UserHandler
func NewUserHandler(store store.User, telemetry *telemetry.Telemetry) *UserHandler {
	return &UserHandler{
		store:     store,
		telemetry: telemetry,
	}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "UserHandler.List")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("List users")

	users, err := h.store.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "UserHandler.Create")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Create users")

	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	created, err := h.store.Create(r.Context(), user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(created)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "UserHandler.Get")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Get users")

	id := r.PathValue("id")
	user, err := h.store.Get(r.Context(), id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "User not found")
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "UserHandler.Update")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("User.update")

	user := &model.User{}
	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedSubscription, err := h.store.Update(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(updatedSubscription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "UserHandler.Delete")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("User.delete")

	id := r.PathValue("id")
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
