package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

func InitMetrics(ctx context.Context) error {
	exp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure())
	if err != nil {
		return err
	}

	provider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
	otel.SetMeterProvider(provider)

	go func() {
		<-ctx.Done()
		if err := provider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down MeterProvider: %v", err)
		}
	}()

	return nil
}
