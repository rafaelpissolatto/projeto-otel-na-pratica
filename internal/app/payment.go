// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	planhttp "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/handler/http"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	storegorm "github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store/gorm"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Payment struct {
	Handler  *planhttp.PaymentHandler
	Store    store.Payment
	natsConn *nats.Conn
	cctx     jetstream.ConsumeContext
	Tracer   trace.Tracer
}

func NewPayment(cfg *config.Payments) (*Payment, error) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(cfg.SQLLite.DSN))
	if err != nil {
		return nil, err
	}
	_ = db.AutoMigrate(&model.Payment{})

	nc, err := nats.Connect(cfg.NATS.Endpoint)
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     cfg.NATS.Stream,
		Subjects: []string{cfg.NATS.Subject},
	})
	if err != nil {
		return nil, err
	}

	stream, err := js.Stream(ctx, cfg.NATS.Stream)
	if err != nil {
		return nil, err
	}

	// this is only relevant for the consumer
	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Name:          cfg.NATS.ConsumerName,
		Durable:       cfg.NATS.ConsumerName,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}

	store := storegorm.NewPaymentStore(db)
	tracer := otel.Tracer("payment")
	pmt := &Payment{
		Handler:  planhttp.NewPaymentHandler(store, js, cfg.NATS.Subject, cfg.SubscriptionsEndpoint),
		Store:    store,
		natsConn: nc,
		Tracer:   tracer,
	}

	pmt.cctx, err = cons.Consume(func(msg jetstream.Msg) {
		pmt.Handler.OnMessage(context.Background(), msg)
	})
	if err != nil {
		return nil, err
	}

	return pmt, nil
}

func (a *Payment) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /payments", a.traceHandler(a.Handler.List))
	mux.HandleFunc("POST /payments", a.traceHandler(a.Handler.Create))
	mux.HandleFunc("GET /payments/{id}", a.traceHandler(a.Handler.Get))
	mux.HandleFunc("PUT /payments/{id}", a.traceHandler(a.Handler.Update))
	mux.HandleFunc("DELETE /payments/{id}", a.traceHandler(a.Handler.Delete))
}

func (a *Payment) Shutdown() error {
	if a.cctx != nil {
		a.cctx.Drain()
	}
	return a.natsConn.Drain()
}

func (a *Payment) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}
