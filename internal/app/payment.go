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
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Payment struct {
	Handler   *planhttp.PaymentHandler
	Store     store.Payment
	natsConn  *nats.Conn
	cctx      jetstream.ConsumeContext
	Telemetry *telemetry.Telemetry
}

func NewPayment(cfg *config.Payments, telemetry *telemetry.Telemetry) (*Payment, error) {
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

	// this is only relevant for testing, etc
	// _, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
	// 	Name:     cfg.NATS.Stream,
	// 	Subjects: []string{cfg.NATS.Subject},
	// })
	// if err != nil {
	// 	return nil, err
	// }

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
	pmt := &Payment{
		Handler:   planhttp.NewPaymentHandler(store, js, cfg.NATS.Subject, cfg.SubscriptionsEndpoint, telemetry),
		Store:     store,
		natsConn:  nc,
		Telemetry: telemetry,
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
	mux.HandleFunc("GET /payments", a.chainHandlers(a.Handler.List))
	mux.HandleFunc("POST /payments", a.chainHandlers(a.Handler.Create))
	mux.HandleFunc("GET /payments/{id}", a.chainHandlers(a.Handler.Get))
	mux.HandleFunc("PUT /payments/{id}", a.chainHandlers(a.Handler.Update))
	mux.HandleFunc("DELETE /payments", a.chainHandlers(a.Handler.Delete))
}

func (a *Payment) chainHandlers(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return a.traceHandler(a.logHandler(a.metricHandler(handlerFunc)))
}

func (a *Payment) Shutdown() error {
	if a.cctx != nil {
		a.cctx.Drain()
	}
	return a.natsConn.Drain()
}

func (a *Payment) traceHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.Telemetry.Tracer.Start(r.Context(), r.Method+" "+r.URL.Path)
		defer span.End()
		handlerFunc(w, r.WithContext(ctx))
	}
}

func (a *Payment) metricHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		counter, err := a.Telemetry.Meter.Int64Counter("payment.requests")
		if err != nil {
			http.Error(w, "failed to create counter", http.StatusInternalServerError)
			return
		}
		counter.Add(r.Context(), 1)
		handlerFunc(w, r)
	}
}

func (a *Payment) logHandler(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// logger.Info(r.Context(), "request", r.Method+" "+r.URL.Path)
		a.Telemetry.Logger.Info("request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
		handlerFunc(w, r)
	}
}
