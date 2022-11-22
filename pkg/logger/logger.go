package logger

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
)

type Logger interface {
	Println(format string, a ...interface{})
	Actionf(format string, a ...interface{})
	Failuref(format string, a ...interface{})
	Generatef(format string, a ...interface{})
	Successf(format string, a ...interface{})
	Waitingf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	L() logr.Logger
}

type CliLogger struct {
	logr.Logger
}

func (l *CliLogger) L() logr.Logger {
	return l.Logger
}

// NewCLILogger returns a wrapped logr that logs to the specified writer
// Note: unless you're doing CLI work, you should use core/logger.New instead
func NewCLILogger(writer io.Writer) Logger {
	return &CliLogger{defaultLogr(writer)}
}

// From wraps a logr instance with the extra emoji generating helpers
func From(logger logr.Logger) Logger {
	return &CliLogger{Logger: logger}
}

func (l *CliLogger) Println(format string, a ...interface{}) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l *CliLogger) Actionf(format string, a ...interface{}) {
	l.Info("► " + fmt.Sprintf(format, a...))
}

func (l *CliLogger) Failuref(format string, a ...interface{}) {
	l.Info("✗ " + fmt.Sprintf(format, a...))
}

func (l *CliLogger) Generatef(format string, a ...interface{}) {
	l.Info("✚ " + fmt.Sprintf(format, a...))
}

func (l *CliLogger) Successf(format string, a ...interface{}) {
	l.Info("✔ " + fmt.Sprintf(format, a...))
}

func (l *CliLogger) Waitingf(format string, a ...interface{}) {
	l.Info("◎ " + fmt.Sprintf(format, a...))
}

func (l *CliLogger) Warningf(format string, a ...interface{}) {
	l.Info("⚠️ " + fmt.Sprintf(format, a...))
}

func defaultLogr(w io.Writer) logr.Logger {
	//jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
	})
	consoleOut := zapcore.Lock(zapcore.AddSync(w))

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
