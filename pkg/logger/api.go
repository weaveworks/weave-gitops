package logger

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

type ApiLogger struct {
	logger logr.Logger
}

func NewApiLogger() Logger {
	zap, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logger := zapr.NewLogger(zap)

	return ApiLogger{
		logger: logger,
	}
}

func (l ApiLogger) Println(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Printf(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Actionf(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Generatef(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Waitingf(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Successf(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Warningf(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Failuref(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Write(p []byte) (n int, err error) {
	l.logger.Info(string(p))
	return len(p), nil
}
