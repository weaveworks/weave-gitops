package gitproviders

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Client
type Client interface {
	GetProvider(repoUrl RepoURL, getAccountType AccountTypeGetter) (GitProvider, error)
}
