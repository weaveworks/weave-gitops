package osys

import (
	"errors"
	"os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Osys
type Osys interface {
	UserHomeDir() (string, error)
	GetGitProviderToken(tokenVarName string) (string, error)
	Getenv(envVar string) string
	LookupEnv(envVar string) (string, bool)
	Setenv(envVar, value string) error
	Unsetenv(envVar string) error
	Exit(code int)
	Stdin() *os.File
	Stdout() *os.File
	Stderr() *os.File
}

type OsysClient struct{}

func New() *OsysClient {
	return &OsysClient{}
}

var _ Osys = &OsysClient{}

func (o *OsysClient) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (o *OsysClient) Getenv(envVar string) string {
	return os.Getenv(envVar)
}

func (o *OsysClient) LookupEnv(envVar string) (string, bool) {
	return os.LookupEnv(envVar)
}

func (o *OsysClient) Setenv(envVar, value string) error {
	return os.Setenv(envVar, value)
}

func (o *OsysClient) Unsetenv(envVar string) error {
	return os.Unsetenv(envVar)
}

// The following three functions are used by both "app add" and "app remove".
// They are here rather than in "utils" so they can use the (potentially mocked)
// local versions of UserHomeDir, LookupEnv, and Stdin and so that they can also
// be mocked (e.g. we might want to mock the private key password handing).

var ErrNoGitProviderTokenSet = errors.New("no git provider token env variable set")

func (o *OsysClient) GetGitProviderToken(tokenVarName string) (string, error) {
	providerToken, found := o.LookupEnv(tokenVarName)

	if !found || providerToken == "" {
		return "", ErrNoGitProviderTokenSet
	}

	return providerToken, nil
}

func (o *OsysClient) Exit(code int) {
	os.Exit(code)
}

func (o *OsysClient) Stdin() *os.File {
	return os.Stdin
}

func (o *OsysClient) Stdout() *os.File {
	return os.Stdout
}

func (o *OsysClient) Stderr() *os.File {
	return os.Stderr
}
