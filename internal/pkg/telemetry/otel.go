package telemetry

import (
	"context"
	"os"

	config "go.opentelemetry.io/contrib/config/v0.3.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Telemetry struct {
	Tracer trace.Tracer
	Meter  metric.Meter
	Logger *zap.Logger
}

// Setup initializes the OpenTelemetry SDK with the provided configuration file
// It returns a shutdown function to clean up resources and an error if any occurs
func Setup(ctx context.Context, configFilePath string) (func(context.Context) error, error) {
	// Read the configuration file
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	// Interpolate the environment variables in the configuration data
	configData = []byte(os.ExpandEnv(string(configData)))

	// Parse the configuration data
	otelConfig, err := config.ParseYAML(configData)
	if err != nil {
		return nil, err
	}

	// Create a new OpenTelemetry SDK instance with the parsed configuration
	sdk, err := config.NewSDK(config.WithContext(ctx), config.WithOpenTelemetryConfiguration(*otelConfig))
	if err != nil {
		return nil, err
	}

	// Set the global propagator, tracer provider, meter provider, and logger provider
	// note: there is bug in the lib that requires the propagation to be set manually here
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetTracerProvider(sdk.TracerProvider())
	otel.SetMeterProvider(sdk.MeterProvider())
	global.SetLoggerProvider(sdk.LoggerProvider())

	// Return the shutdown function and nil error
	return sdk.Shutdown, nil
}
