package osys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	cryptossh "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Osys
type Osys interface {
	UserHomeDir() (string, error)
	SelectAuthMethod(privateKeyPath string) (ssh.AuthMethod, error)
	GetGitProviderToken() (string, error)
	Getenv(envVar string) string
	LookupEnv(envVar string) (string, bool)
	Setenv(envVar, value string) error
	Exit(code int)
	Stdin() *os.File
	Stdout() *os.File
	Stderr() *os.File
}

const (
	SSHAuthSock = "SSH_AUTH_SOCK"
)

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

func (o *OsysClient) SelectAuthMethod(privateKeyPath string) (ssh.AuthMethod, error) {
	var (
		authMethod ssh.AuthMethod
		authErr    error
	)

	switch {
	case strings.HasPrefix(privateKeyPath, "~/"):
		dir, err := o.UserHomeDir()
		if err != nil {
			return nil, err
		}
		authMethod, authErr = o.authMethodFromKeyFile(filepath.Join(dir, privateKeyPath[2:]))
	case privateKeyPath != "":
		authMethod, authErr = o.authMethodFromKeyFile(privateKeyPath)
	default: // no private key given, try ssh-agent or find a likely key file
		authMethod, authErr = o.NewSshAgentOrFindKeyFile()
	}

	if authErr != nil {
		return nil, authErr
	}

	return authMethod, nil
}

func (o *OsysClient) authMethodFromKeyFile(privateKeyFile string) (*ssh.PublicKeys, error) {
	authMethod, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
	if err != nil {
		fmt.Printf("Enter passphrase for key '%s': ", privateKeyFile)
		pw, err := term.ReadPassword(int(o.Stdin().Fd()))
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("failed reading ssh key password: %w", err)
		}

		authMethod, err = ssh.NewPublicKeysFromFile("git", privateKeyFile, string(pw))
		if err != nil {
			return nil, fmt.Errorf("failed reading ssh keys: %w", err)
		}
	}
	return authMethod, nil
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

// SshAgentOrFindKeyFile implements ssh.AuthMethod by attempting to use
// SSH agent, and failing that, trying standard key locations.
type SshAgentOrFindKeyFile struct {
	ssh.HostKeyCallbackHelper
	agentAuthMethod *ssh.PublicKeysCallback
	oclient         *OsysClient
}

func (o *OsysClient) NewSshAgentOrFindKeyFile() (*SshAgentOrFindKeyFile, error) {
	auth := &SshAgentOrFindKeyFile{oclient: o}
	if o.Getenv(SSHAuthSock) != "" {
		a, err := ssh.NewSSHAgentAuth("") // empty means figure it out for yourself
		if err != nil {
			return nil, err
		}
		auth.agentAuthMethod = a
	}
	return auth, nil
}

func (*SshAgentOrFindKeyFile) Name() string {
	return "ssh-agent-or-find-key-file"
}

func (a *SshAgentOrFindKeyFile) String() string {
	return a.Name()
}

func (a *SshAgentOrFindKeyFile) ClientConfig() (*cryptossh.ClientConfig, error) {
	var auths []cryptossh.AuthMethod
	if a.agentAuthMethod != nil {
		auths = append(auths, cryptossh.PublicKeysCallback(a.agentAuthMethod.Callback))
	}
	auths = append(auths, cryptossh.PublicKeysCallback(a.oclient.signersFromUsualFiles))
	return a.SetHostKeyCallback(&cryptossh.ClientConfig{
		User: "git",
		Auth: auths,
	})
}

func (o *OsysClient) signersFromUsualFiles() ([]cryptossh.Signer, error) {
	pk, err := o.findPrivateKeyFile()
	if err != nil {
		return nil, err
	}
	auth, err := o.authMethodFromKeyFile(pk)
	if err != nil {
		return nil, err
	}
	return []cryptossh.Signer{auth.Signer}, nil
}
