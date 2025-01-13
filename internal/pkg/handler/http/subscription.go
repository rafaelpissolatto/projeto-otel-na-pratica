// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// SubscriptionHandler is an HTTP handler that performs CRUD operations for model.Subscription using a store.Subscription
type SubscriptionHandler struct {
	store         store.Subscription
	usersEndpoint string
	plansEndpoint string
	telemetry     *telemetry.Telemetry
}

// NewSubscriptionHandler returns a new SubscriptionHandler
func NewSubscriptionHandler(store store.Subscription, usersEndpoint string, plansEndpoint string, telemetry *telemetry.Telemetry) *SubscriptionHandler {
	return &SubscriptionHandler{
		store:         store,
		usersEndpoint: usersEndpoint,
		plansEndpoint: plansEndpoint,
		telemetry:     telemetry,
	}
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "SubscriptionHandler.List")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("List subscriptions")

	if r.Method != http.MethodGet {
		span.RecordError(errors.New("method not allowed"))
		span.SetStatus(codes.Error, "Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	subscriptions, err := h.store.List(r.Context())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		h.telemetry.Logger.Error("Failed to list subscriptions", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(subscriptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		h.telemetry.Logger.Error("Failed to encode subscriptions", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "SubscriptionHandler.Create")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Create subscriptions")

	subscription := &model.Subscription{}
	if err := json.NewDecoder(r.Body).Decode(subscription); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		h.telemetry.Logger.Error("Invalid request payload", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// verify the user exists
	{
		user, _ := http.Get(h.usersEndpoint + "/" + subscription.UserID)
		if user.StatusCode != http.StatusOK {
			span.RecordError(errors.New("user not found"))
			span.SetStatus(codes.Error, "User not found")
			h.telemetry.Logger.Error("User not found", zap.String("user_id", subscription.UserID))
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		}
		defer user.Body.Close()
	}

	// verify the plan exists
	{
		plan, _ := http.Get(h.plansEndpoint + "/" + subscription.PlanID)
		if plan.StatusCode != http.StatusOK {
			span.RecordError(errors.New("plan not found"))
			span.SetStatus(codes.Error, "Plan not found")
			h.telemetry.Logger.Error("Plan not found", zap.String("plan_id", subscription.PlanID))
			http.Error(w, "Plan not found", http.StatusBadRequest)
			return
		}
		defer plan.Body.Close()
	}

	created, err := h.store.Create(r.Context(), subscription)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		h.telemetry.Logger.Error("Failed to create subscription", zap.Error(err))
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

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "SubscriptionHandler.Get")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Get subscriptions")

	id := r.PathValue("id")
	subscription, err := h.store.Get(r.Context(), id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if subscription == nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Subscription not found")
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(subscription)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "SubscriptionHandler.Update")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Update subscriptions")

	subscription := &model.Subscription{}
	if err := json.NewDecoder(r.Body).Decode(subscription); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	updatedSubscription, err := h.store.Update(r.Context(), subscription)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(updatedSubscription)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "SubscriptionHandler.Delete")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Delete subscriptions")

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
