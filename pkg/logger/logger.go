package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logr.Logger
}

// NewCLILogger returns a wrapped logr that logs to the specified writer
// Note: unless you're doing CLI work, you should use core/logger.New instead
func NewCLILogger(writer io.Writer) Logger {
	return Logger{defaultLogr()}
}

// From wraps a logr instance with the extra emoji generating helpers
func From(logger logr.Logger) Logger {
	return Logger{Logger: logger}
}

func (l Logger) Println(format string, a ...interface{}) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l Logger) Actionf(format string, a ...interface{}) {
	l.Info("► " + fmt.Sprintf(format, a...))
}

func (l Logger) Failuref(format string, a ...interface{}) {
	l.Info("✗ " + fmt.Sprintf(format, a...))
}

func (l Logger) Generatef(format string, a ...interface{}) {
	l.Info("✚ " + fmt.Sprintf(format, a...))
}

func (l Logger) Successf(format string, a ...interface{}) {
	l.Info("✔ " + fmt.Sprintf(format, a...))
}

func (l Logger) Waitingf(format string, a ...interface{}) {
	l.Info("◎ " + fmt.Sprintf(format, a...))
}

func (l Logger) Warningf(format string, a ...interface{}) {
	l.Info("⚠️ " + fmt.Sprintf(format, a...))
}

func defaultLogr() logr.Logger {
	//jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
	})
	consoleOut := zapcore.Lock(os.Stdout)

	// Should point into the cluster
	//clusterOut := zapcore.Lock(os.Stderr)

	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleOut, level),
		//zapcore.NewCore(jsonEncoder, clusterOut, level),
	)

	logger := zap.New(core)

	return zapr.NewLogger(logger)
}
