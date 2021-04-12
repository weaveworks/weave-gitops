package test

// Runs basic WeGO operations against a kind cluster.

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/status"
)

// Run core operations and check status
func TestCoreOperations(t *testing.T) {
	savedHome := os.Getenv("HOME")

	err := os.Setenv("HOME", "/iewojfoiwejfoiwjfwoijfewj")
	require.NoError(t, err)
	require.Equal(t, status.GetClusterStatus(), status.Unknown)

	err = os.Setenv("HOME", savedHome)
	require.NoError(t, err)
	require.Equal(t, status.GetClusterStatus(), status.Unmodified)

	out, err := fluxops.CallFlux("bootstrap github -- --owner=jrryjcksn --repository=fleet-infra --branch=main --private=false --personal=true")
	require.NoError(t, err)
}
