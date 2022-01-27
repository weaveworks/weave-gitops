package vendorfakes

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate -o fakegitprovider -fake-name Client github.com/fluxcd/go-git-providers/gitprovider.Client
//counterfeiter:generate -o fakegitprovider -fake-name OrgRepositoriesClient github.com/fluxcd/go-git-providers/gitprovider.OrgRepositoriesClient
//counterfeiter:generate -o fakegitprovider -fake-name UserRepositoriesClient github.com/fluxcd/go-git-providers/gitprovider.UserRepositoriesClient
//counterfeiter:generate -o fakegitprovider -fake-name DeployKeyClient github.com/fluxcd/go-git-providers/gitprovider.DeployKeyClient
//counterfeiter:generate -o fakegitprovider -fake-name OrgRepository github.com/fluxcd/go-git-providers/gitprovider.OrgRepository
//counterfeiter:generate -o fakegitprovider -fake-name UserRepository github.com/fluxcd/go-git-providers/gitprovider.UserRepository
//counterfeiter:generate -o fakegitprovider -fake-name BranchClient github.com/fluxcd/go-git-providers/gitprovider.BranchClient
//counterfeiter:generate -o fakegitprovider -fake-name CommitClient github.com/fluxcd/go-git-providers/gitprovider.CommitClient
//counterfeiter:generate -o fakegitprovider -fake-name PullRequestClient github.com/fluxcd/go-git-providers/gitprovider.PullRequestClient
//counterfeiter:generate -o fakegitprovider -fake-name Commit github.com/fluxcd/go-git-providers/gitprovider.Commit
//counterfeiter:generate -o fakegitprovider -fake-name FileClient github.com/fluxcd/go-git-providers/gitprovider.FileClient
//counterfeiter:generate -o fakegitprovider -fake-name PullRequest github.com/fluxcd/go-git-providers/gitprovider.PullRequest

//counterfeiter:generate -o fakelogr -fake-name Logger github.com/go-logr/logr.Logger

//counterfeiter:generate -o fakehttp -fake-name Handler net/http.Handler
//counterfeiter:generate -o fakehttp -fake-name RoundTripper net/http.RoundTripper
