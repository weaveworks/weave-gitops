package internal

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type ApiLogger struct {
	logger logr.Logger
}

func NewApiLogger(l logr.Logger) logger.Logger {
	return ApiLogger{
		logger: l,
	}
}

func (l ApiLogger) Println(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Printf(format string, a ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, a...))
}

func (l ApiLogger) Infow(msg string, kv ...interface{}) {
	l.logger.Info(msg, kv...)
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
