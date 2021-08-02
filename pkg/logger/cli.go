package logger

import (
	"fmt"
	"io"
)

type CLILogger struct {
	stdout io.Writer
}

func NewCLILogger(writer io.Writer) Logger {
	return CLILogger{
		stdout: writer,
	}
}

func (l CLILogger) Println(format string, a ...interface{}) {
	fmt.Fprintln(l.stdout, fmt.Sprintf(format, a...))
}

func (l CLILogger) Printf(format string, a ...interface{}) {
	fmt.Fprintf(l.stdout, format, a...)
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
