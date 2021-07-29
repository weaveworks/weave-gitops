package osys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/term"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Osys
type Osys interface {
	UserHomeDir() (string, error)
	CanonicalPrivateKeyFile(candidate string) (string, error)
	RetrievePublicKeyFromFile(filename string) (*ssh.PublicKeys, error)
	GetGitProviderToken() (string, error)
	Getenv(envVar string) string
	LookupEnv(envVar string) (string, bool)
	Setenv(envVar, value string) error
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

// The following three functions are used by both "app add" and "app remove".
// They are here rather than in "utils" so they can use the (potentially mocked)
// local versions of UserHomeDir, LookupEnv, and Stdin and so that they can also
// be mocked (e.g. we might want to mock the private key password handing).

func (o *OsysClient) GetGitProviderToken() (string, error) {
	providerToken, found := o.LookupEnv("GITHUB_TOKEN")
	if !found {
		return "", fmt.Errorf("GITHUB_TOKEN not set in environment")
	}

	return providerToken, nil
}

func (o *OsysClient) CanonicalPrivateKeyFile(candidate string) (string, error) {
	if strings.HasPrefix(candidate, "~/") {
		dir, err := o.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(dir, candidate[2:]), nil
	} else if candidate == "" {
		privateKey, err := o.findPrivateKeyFile()
		if err != nil {
			return "", err
		}

		return privateKey, nil
	} else {
		return candidate, nil
	}
}

func (o *OsysClient) RetrievePublicKeyFromFile(filename string) (*ssh.PublicKeys, error) {
	authMethod, err := ssh.NewPublicKeysFromFile("git", filename, "")
	if err != nil {
		fmt.Print("Private Key Password: ")
		pw, err := term.ReadPassword(int(o.Stdin().Fd()))
		if err != nil {
			return nil, fmt.Errorf("failed reading ssh key password: %w", err)
		}

		authMethod, err = ssh.NewPublicKeysFromFile("git", filename, string(pw))
		if err != nil {
			return nil, fmt.Errorf("failed reading ssh keys: %w", err)
		}
	}

	return authMethod, nil
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

func (o *OsysClient) findPrivateKeyFile() (string, error) {
	dir, err := o.UserHomeDir()
	if err != nil {
		return "", err
	}

	modernFilePath := filepath.Join(dir, ".ssh", "id_ed25519")
	if utils.Exists(modernFilePath) {
		return modernFilePath, nil
	}

	legacyFilePath := filepath.Join(dir, ".ssh", "id_rsa")
	if utils.Exists(legacyFilePath) {
		return legacyFilePath, nil
	}

	return "", fmt.Errorf("could not locate ssh key file")
}
