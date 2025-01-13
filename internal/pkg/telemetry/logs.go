package telemetry

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
)

var (
	logger *logrus.Logger
)

func InitLogs(ctx context.Context) error {
	logger = logrus.New()

	// log.SetFormatter(&log.JSONFormatter{}) //TODO: see whats the team prefer to use
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})	

	logrus.AddHook(otellogrus.NewHook(otellogrus.WithLevels(
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
	)))

	return nil
}

// GetLogger allows accessing the initialized logger
func GetLogger() *logrus.Logger {
	if logger == nil {
		panic(errors.New("logger not initialized"))
	}

	return logger
}
