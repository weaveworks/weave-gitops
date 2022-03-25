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

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Logger
type Logger interface {
	Println(format string, a ...interface{})
	Printf(format string, a ...interface{})
	Infow(msg string, kv ...interface{})
	Actionf(format string, a ...interface{})
	Generatef(format string, a ...interface{})
	Waitingf(format string, a ...interface{})
	Successf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Failuref(format string, a ...interface{})
	Write(p []byte) (n int, err error) // added to satisfy io.Writer interface
}
