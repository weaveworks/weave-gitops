package status

import (
	"fmt"
	"strings"
	"testing"

	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/stretchr/testify/require"
)

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

	override.WithOverrides(func() override.Result {
		name, err := GetClusterName()
		require.NoError(t, err)
		require.Equal(t, name, "kind-wego-demo")
		return override.Result{}
	}, utils.OverrideBehavior(utils.CallCommandSeparatingOutputStreamsOp,
		func(args ...interface{}) ([]byte, []byte, error) {
			case0Kubectl := `kubectl config current-context`
			switch (args[0]).(string) {
			case case0Kubectl:
				return []byte("kind-wego-demo"), []byte(""), nil
			default:
				return nil, nil, fmt.Errorf("arguments not expected %s", args)
			}
		}),
	)

	override.WithOverrides(func() override.Result {
		_, err := GetClusterName()
		require.Error(t, err)
		return override.Result{}
	}, utils.OverrideBehavior(utils.CallCommandSeparatingOutputStreamsOp,
		func(args ...interface{}) ([]byte, []byte, error) {
			case0Kubectl := `kubectl config current-context`
			switch (args[0]).(string) {
			case case0Kubectl:
				return []byte("wrong-name"), []byte(""), fmt.Errorf("error here")
			default:
				return nil, nil, fmt.Errorf("arguments not expected %s", args)
			}
		}),
	)

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
