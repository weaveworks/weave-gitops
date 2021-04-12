package test

// Runs basic WeGO operations against a kind cluster.

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/cluster-api-provider-existinginfra/apis/cluster.weave.works/v1alpha3"
	capeios "github.com/weaveworks/cluster-api-provider-existinginfra/pkg/apis/wksprovider/machine/os"
	clientcmdv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/yaml"
)

// Run core operations and check status
func TestCoreOperations(t *testing.T) {
	savedHome := os.Getenv("HOME")
	err := os.Setenv("HOME", "/iewojfoiwejfoiwjfwoijfewj")
	require.NoError(t, err)
	require.Equal(t, status.GetClusterStatus(), status.Unknown)
	err := os.Setenv("HOME", savedHome)
	require.NoError(t, err)
	require.Equal(t, status.GetClusterStatus(), status.Unmodified)
	callFlux("bootstrap github --owner=$GITHUB_USER --repository=fleet-infra --branch=main --path=./clusters/my-cluster --personal")
}
