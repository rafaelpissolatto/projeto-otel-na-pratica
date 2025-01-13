// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/app"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
	"github.com/mattn/go-colorable"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/log/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// main is the entry point of the all-in-one service. It sets up telemetry, logging, and configuration,
// and starts various services including user, plan, payment, and subscription services. It also starts
// both HTTP and gRPC servers to handle incoming requests. The function has a shutdown of services
// and logs any errors encountered during the setup and runtime
func main() {
	configFlag := flag.String("config", "", "path to the config file")
	otelConfigFlag := flag.String("otel", "./configs/otel/otel.yaml", "path to the OTel config file")
	flag.Parse()

	// Setup telemetry with the provided configuration file
	closer, err := telemetry.Setup(context.Background(), *otelConfigFlag)
	if err != nil {
		fmt.Printf("failed to setup telemetry: %v\n", err)
	}
	defer closer(context.Background())

	// Logging setup
	// core := zapcore.NewTee(
	// 	zapcore.NewCore(zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
	// 		zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
	// 	otelzap.NewCore("all-in-one", otelzap.WithLoggerProvider(global.GetLoggerProvider())),
	// )

	encoderLogger := zap.NewDevelopmentEncoderConfig()
	encoderLogger.EncodeLevel = zapcore.CapitalColorLevelEncoder
	coreLogger := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderLogger),
			zapcore.AddSync(colorable.NewColorableStdout()),
			zapcore.DebugLevel,
		),
		otelzap.NewCore("all-in-one", otelzap.WithLoggerProvider(global.GetLoggerProvider())),
	)
	logger := zap.New(coreLogger)

	ctx, span := otel.Tracer("all-in-one").Start(context.Background(), "main")
	defer span.End()

	logger.Debug("starting the all-in-one service")
	span.AddEvent("starting the all-in-one service")
	c, err := config.LoadConfig(*configFlag)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Fatal("failed to load the config", zap.Error(err))
	}

	// Create a new HTTP server mux
	mux := http.NewServeMux()

	// starts the gRPC server
	lis, err := net.Listen("tcp", c.Server.Endpoint.GRPC)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// Create a new gRPC server
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	{
		logger.Info("starting the user service")
		span.AddEvent("starting the user service")
		userTelemetry := &telemetry.Telemetry{
			Tracer: otel.Tracer("user"),
			Meter:  otel.Meter("user"),
			Logger: logger,
		}
		a := app.NewUser(ctx, &c.Users, userTelemetry)
		a.RegisterRoutes(mux)
	}

	{
		logger.Info("starting the plan service")
		span.AddEvent("starting the plan service")
		planTelemetry := &telemetry.Telemetry{
			Tracer: otel.Tracer("plan"),
			Meter:  otel.Meter("plan"),
			Logger: logger,
		}
		a := app.NewPlan(&c.Plans, planTelemetry)
		a.RegisterRoutes(mux, grpcServer)
	}

	{
		span.AddEvent("starting the payment service")
		logger.Info("starting the payment service")
		paymentTelemetry := &telemetry.Telemetry{
			Tracer: otel.Tracer("payment"),
			Meter:  otel.Meter("payment"),
			Logger: logger,
		}
		a, err := app.NewPayment(&c.Payments, paymentTelemetry)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Fatal("failed to create the payment service", zap.Error(err))
		}
		a.RegisterRoutes(mux)
		defer func() {
			err = a.Shutdown()
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				logger.Fatal("failed to shutdown the payment service", zap.Error(err))
			}
		}()
	}

	{
		logger.Info("starting the subscription service")
		span.AddEvent("starting the subscription service")
		subscriptionTelemetry := &telemetry.Telemetry{
			Tracer: otel.Tracer("subscription"),
			Meter:  otel.Meter("subscription"),
			Logger: logger,
		}
		a := app.NewSubscription(&c.Subscriptions, subscriptionTelemetry)
		a.RegisterRoutes(mux)
	}

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	// span ends will be called in the defer function
	// span.End()

	// Start the HTTP server
	err = http.ListenAndServe(c.Server.Endpoint.HTTP, mux)
	if err != nil && err != http.ErrServerClosed {
		logger.Error("failed to serve", zap.Error(err))
	}

	logger.Info("stopping the all-in-one service")
}
