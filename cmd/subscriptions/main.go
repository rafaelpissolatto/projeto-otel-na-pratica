// Copyright Dose de Telemetria GmbH
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/app"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/config"
	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/pkg/telemetry"
)

func main() {
	configFlag := flag.String("config", "", "path to the config file")
	flag.Parse()
	ctx := context.Background()

	// init telemetry
	telemetry.InitTelemetry(ctx)
	telemetry.InitMetrics(ctx)
	telemetry.InitLogs(ctx)

	// Retrieve Logger
	logger := telemetry.GetLogger()

	c, _ := config.LoadConfig(*configFlag)

	a := app.NewSubscription(&c.Subscriptions)
	a.RegisterRoutes(http.DefaultServeMux)

	logger.Infof("starting server on %s", c.Server.Endpoint.HTTP)
	_ = http.ListenAndServe(c.Server.Endpoint.HTTP, http.DefaultServeMux)
}
