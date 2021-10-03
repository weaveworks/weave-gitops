package kube

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api/v1"
	kyaml "sigs.k8s.io/yaml"
)

func TestInitialContexts(t *testing.T) {

	InClusterConfig = func() (*rest.Config, error) { return nil, rest.ErrNotInCluster }
	origkc := os.Getenv("KUBECONFIG")
	defer os.Setenv("KUBECONFIG", origkc)
	dir, err := ioutil.TempDir("", "kube-http-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	tests := []struct {
		name        string
		clusterName string
	}{
		{"foo", "foo"},
		{"weave-upa-admin@weave-upa", "weave-upa"},
	}
	for _, test := range tests {
		createKubeconfig(t, test.name, test.clusterName, dir)
		curContext, clusterName, err := initialContexts(clientcmd.NewDefaultClientConfigLoadingRules())
		assert.NoError(t, err)
		assert.Equal(t, test.clusterName, clusterName)
		assert.Equal(t, test.name, curContext)
	}

}
func createKubeconfig(t *testing.T, name, clusterName, dir string) {

	f, err := ioutil.TempFile(dir, "test.kubeconfig")
	assert.NoError(t, err)
	c := clientcmdapi.Config{}
	c.CurrentContext = name
	c.APIVersion = "v1"
	c.Kind = "Config"
	c.Contexts = append(c.Contexts, clientcmdapi.NamedContext{Name: name, Context: clientcmdapi.Context{Cluster: clusterName}})
	c.Clusters = append(c.Clusters, clientcmdapi.NamedCluster{Name: clusterName, Cluster: clientcmdapi.Cluster{Server: "127.0.0.1"}})
	kubeconfig, err := kyaml.Marshal(&c)
	assert.NoError(t, err)
	_, err = f.Write(kubeconfig)
	assert.NoError(t, err)
	f.Close()
	os.Setenv("KUBECONFIG", f.Name())
}
