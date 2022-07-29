package vendorfakes

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakelogr -fake-name LogSink github.com/go-logr/logr.LogSink
