package status

import (
	"fmt"
	"strings"
	"testing"

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
