package logger

import (
	"fmt"
	"io"
	"log"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CLILogger struct {
	stdout io.Writer
}

func NewCLILogger(writer io.Writer) logger.Logger {
	return CLILogger{
		stdout: writer,
	}
}

func Logr() logr.Logger {
	l, err := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "msg",
		},
		OutputPaths: []string{"stdout"},
	}.Build()
	if err != nil {
		log.Fatalf("Couldn't set up logger: %v", err)
	}

	return zapr.NewLogger(l)
}

func (l CLILogger) Println(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, fmt.Sprintf(format, a...))
}

func (l CLILogger) Printf(format string, a ...interface{}) {
	fmt.Fprintf(l.stdout, format, a...)
}

func (l CLILogger) Infow(msg string, kv ...interface{}) {
	var s, k string
	for _, v := range kv {
		if k == "" {
			k = fmt.Sprintf("%s", v)
			continue
		}

		s += fmt.Sprintf(" %s=%v ", k, v)
		k = ""
	}

	fmt.Fprintf(l.stdout, msg+" - %s\n", s)
}

func (l CLILogger) Actionf(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, `►`, fmt.Sprintf(format, a...))
}

func (l CLILogger) Generatef(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, `✚`, fmt.Sprintf(format, a...))
}

func (l CLILogger) Waitingf(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, `◎`, fmt.Sprintf(format, a...))
}

func (l CLILogger) Successf(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, `✔`, fmt.Sprintf(format, a...))
}

func (l CLILogger) Warningf(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, `⚠️`, fmt.Sprintf(format, a...))
}

func (l CLILogger) Failuref(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, `✗`, fmt.Sprintf(format, a...))
}

func (l CLILogger) Write(p []byte) (n int, err error) {
	n, err = l.stdout.Write(p)
	if err != nil {
		return n, err
	}

	if n != len(p) {
		return n, io.ErrShortWrite
	}

	return len(p), nil
}
