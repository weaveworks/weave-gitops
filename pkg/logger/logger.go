/*
Copyright 2020 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logger

import (
	"fmt"
	"io"
)

type Logger interface {
	Println(format string, a ...interface{})
	Printf(format string, a ...interface{})
	Actionf(format string, a ...interface{})
	Generatef(format string, a ...interface{})
	Waitingf(format string, a ...interface{})
	Successf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Failuref(format string, a ...interface{})
}

type CLILogger struct {
	stdout io.Writer
}

func New(writer io.Writer) Logger {
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
