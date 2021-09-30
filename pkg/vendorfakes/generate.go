package vendorfakes

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakegitprovider -fake-name Client github.com/fluxcd/go-git-providers/gitprovider.Client
//counterfeiter:generate -o fakelogr -fake-name Logger github.com/go-logr/logr.Logger
//counterfeiter:generate -o fakehttp -fake-name Handler net/http.Handler
//counterfeiter:generate -o fakehttp -fake-name RoundTripper net/http.RoundTripper
