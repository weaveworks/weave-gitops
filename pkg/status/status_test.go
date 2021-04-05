package status

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClusterStatus(t *testing.T) {
	lookupHandler = func(args string) error {
		return fmt.Errorf("Failed calling kubectl get %s", args)
	}
	require.Equal(t, GetClusterStatus(), Unknown)

	lookupHandler = func(args string) error {
		if !strings.HasPrefix(args, "deployment coredns") {
			return fmt.Errorf("Failed calling kubectl get %s", args)
		}
		return nil
	}
	require.Equal(t, GetClusterStatus(), Unmodified)

	lookupHandler = func(args string) error {
		if !strings.HasPrefix(args, "customresourcedefinition") {
			return fmt.Errorf("Failed calling kubectl get %s", args)
		}
		return nil
	}
	require.Equal(t, GetClusterStatus(), FluxInstalled)

	lookupHandler = func(args string) error {
		if !strings.HasPrefix(args, "deployment wego-controller") {
			return fmt.Errorf("Failed calling kubectl get %s", args)
		}
		return nil
	}

	require.Equal(t, GetClusterStatus(), WeGOInstalled)
}
