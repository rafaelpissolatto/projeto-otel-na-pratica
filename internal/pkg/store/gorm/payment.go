// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package gorm

import (
	"context"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/model"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/store"
	"gorm.io/gorm"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Payment struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewPaymentStore(db *gorm.DB) store.Payment {
	tracer := otel.Tracer("payment store gorm")
	return &Payment{
		db:     db,
		tracer: tracer,
	}
}

func (p *Payment) Get(ctx context.Context, id string) (*model.Payment, error) {
	_, span := p.tracer.Start(ctx, "PaymentStore.Get")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Get payment store")

	ret := &model.Payment{}
	_ = p.db.WithContext(ctx).Model(ret).First(&ret, "id = ?", id)
	return ret, nil
}

func (p *Payment) Create(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	_, span := p.tracer.Start(ctx, "PaymentStore.Create")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Create payment store")

	res := p.db.WithContext(ctx).Create(&payment)
	return payment, res.Error
}

func (p *Payment) Update(ctx context.Context, payment *model.Payment) (*model.Payment, error) {
	_, span := p.tracer.Start(ctx, "PaymentStore.Update")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Update payment store")

	res := p.db.WithContext(ctx).Save(&payment)
	return payment, res.Error
}

func (p *Payment) Delete(ctx context.Context, id string) error {
	_, span := p.tracer.Start(ctx, "PaymentStore.Delete")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("Delete payment store")

	_ = p.db.WithContext(ctx).Delete(&model.Payment{}, "id = ?", id)
	return nil
}

func (p *Payment) List(ctx context.Context) ([]*model.Payment, error) {
	_, span := p.tracer.Start(ctx, "PaymentStore.List")
	defer span.End()
	span.SetAttributes()
	span.AddEvent("List payment store")

	var ret []*model.Payment
	_ = p.db.WithContext(ctx).Find(&ret)
	return ret, nil
}
