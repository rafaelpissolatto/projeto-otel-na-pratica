package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitTraces(ctx context.Context) error {
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
	if err != nil {
		return err
	}

	provider := trace.NewTracerProvider(trace.WithBatcher(exp))
	otel.SetTracerProvider(trace.NewTracerProvider(trace.WithBatcher(exp)))
	go func() {
		<-ctx.Done()
		if err := provider.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down TracingProvider: %v", err)
		}
	}()
	return nil
}
