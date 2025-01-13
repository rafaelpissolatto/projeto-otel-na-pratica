package telemetry

import "context"

func InitTelemetry(ctx context.Context) {
	InitTraces(ctx)
	InitMetrics(ctx)
	InitLogs(ctx)
}
