package status

import (
	"fmt"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const kubeconfig = `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: stuff
    server: https://127.0.0.1:46677
  name: kind-wego-demo
contexts:
- context:
    cluster: kind-wego-demo
    user: kind-wego-demo
  name: kind-wego-demo
current-context: kind-wego-demo
kind: Config
preferences: {}
users:
- name: kind-wego-demo
  user:
    client-certificate-data: more stuff
    client-key-data: yet more stuff
`

func TestClusterStatus(t *testing.T) {
	lookupHandler = fail
	require.Equal(t, GetClusterStatus(), Unknown)

	lookupHandler = handle("deployment coredns")
	require.Equal(t, GetClusterStatus(), Unmodified)

	lookupHandler = handle("customresourcedefinition")
	require.Equal(t, GetClusterStatus(), FluxInstalled)

	lookupHandler = handle("deployment wego-controller")
	require.Equal(t, GetClusterStatus(), WeGOInstalled)
}

func TestGetClusterName(t *testing.T) {
	tmpPath, err := ioutil.TempDir("", "tmp-dir")
	require.NoError(t, err)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	defer func() {
		if err := os.Setenv("HOME", home); err != nil {
			require.FailNow(t, "Failed to reset home directory")
		}
	}()
	require.NoError(t, os.Setenv("HOME", tmpPath))
	configDirPath := filepath.Join(tmpPath, ".kube")
	require.NoError(t, os.MkdirAll(configDirPath, 0755))
	require.NoError(t, ioutil.WriteFile(filepath.Join(configDirPath, "config"), []byte(kubeconfig), 0644))
	name, err := utils.GetClusterName()
	require.NoError(t, err)
	require.Equal(t, name, "kind-wego-demo")
}

func handle(prefix string) func(args string) error {
	return func(args string) error {
		if !strings.HasPrefix(args, prefix) {
			return fail(args)
		}
		return nil
	}
}

func fail(args string) error {
	return fmt.Errorf("Failed calling kubectl get %s", args)
}
