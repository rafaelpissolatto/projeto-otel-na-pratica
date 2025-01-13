// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// PaymentHandler is an HTTP handler that performs CRUD operations for model.Payment using a store.Payment
type PaymentHandler struct {
	store                 store.Payment
	js                    jetstream.JetStream
	jsSubject             string
	subscriptionsEndpoint string
	telemetry             *telemetry.Telemetry
}

// NewPaymentHandler returns a new PaymentHandler
func NewPaymentHandler(store store.Payment, js jetstream.JetStream, jsSubject string, subscriptionsEndpoint string, telemetry *telemetry.Telemetry) *PaymentHandler {
	return &PaymentHandler{
		store:                 store,
		js:                    js,
		jsSubject:             jsSubject,
		subscriptionsEndpoint: subscriptionsEndpoint,
		telemetry:             telemetry,
	}
}

func (h *PaymentHandler) List(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "PaymentHandler.List")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("List payments")

	payments, err := h.store.List(r.Context())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(payments)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "PaymentHandler.Create")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Create payments")

	var payment model.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		h.telemetry.Logger.Error("Invalid request payload", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Check if subscription exists
	sub, _ := http.Get(h.subscriptionsEndpoint + "/" + payment.SubscriptionID)
	if sub.StatusCode != http.StatusOK {
		span.RecordError(errors.New("subscription not found"))
		span.SetStatus(codes.Error, "subscription not found")
		http.Error(w, "Subscription not found", http.StatusBadRequest)
		return
	}
	defer sub.Body.Close()

	payload, err := json.Marshal(payment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Propagate the context
	// TODO: not working, review this
	hs := nats.Header{
		"X-Request-ID": []string{span.SpanContext().TraceID().String()},
	}
	otel.GetTextMapPropagator().Inject(r.Context(), carrier{hs})

	// Publish the payment to the JetStream
	_, err = h.js.PublishMsgAsync(&nats.Msg{
		Subject: h.jsSubject,
		Data:    payload,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(payment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "PaymentHandler.Get")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Get payments")

	id := r.PathValue("id")
	payment, err := h.store.Get(r.Context(), id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if payment == nil {
		err := errors.New("payment not found")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(payment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Update(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "PaymentHandler.Update")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Update payments")

	payment := &model.Payment{}
	if err := json.NewDecoder(r.Body).Decode(payment); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err := h.store.Update(r.Context(), payment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(payment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_, span := h.telemetry.Tracer.Start(r.Context(), "PaymentHandler.Delete")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Delete payments")

	id := r.PathValue("id")
	err := h.store.Delete(r.Context(), id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PaymentHandler) OnMessage(ctx context.Context, msg jetstream.Msg) {
	_, span := h.telemetry.Tracer.Start(ctx, "PaymentHandler.OnMessage")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("OnMessage payments")

	payment := &model.Payment{}
	err := json.Unmarshal(msg.Data(), payment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	_, err = h.store.Create(context.Background(), payment)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	_ = msg.Ack()
}

// carrier is a type that implements the TextMapReader interface
// note: this is custom impl to allow use the TextMapPropagator to inject the trace context into the NATS message
type carrier struct {
	nats.Header
}

// Get implements the TextMapReader interface
func (c carrier) Keys() []string {
	keys := make([]string, 0, len(c.Header))
	for k := range c.Header {
		keys = append(keys, k)
	}
	return keys
}
