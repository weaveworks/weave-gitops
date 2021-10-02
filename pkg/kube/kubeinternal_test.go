package kube

import (
	"fmt"
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
	dir, err := ioutil.TempDir("", "kube-http-test-")
	assert.NoError(t, err)

	origkc := os.Getenv("KUBECONFIG")
	f, err := ioutil.TempFile(dir, "test.kubeconfig")
	assert.NoError(t, err)
	c := clientcmdapi.Config{}
	c.CurrentContext = "foo"
	c.APIVersion = "v1"
	c.Kind = "Config"
	c.Contexts = append(c.Contexts, clientcmdapi.NamedContext{Name: "foo", Context: clientcmdapi.Context{Cluster: "mycluster"}})
	c.Clusters = append(c.Clusters, clientcmdapi.NamedCluster{Name: "mycluster", Cluster: clientcmdapi.Cluster{Server: "127.0.0.1"}})
	fmt.Printf("----- %+v\n", c)
	kubeconfig, err := kyaml.Marshal(&c)
	assert.NoError(t, err)
	_, err = f.Write(kubeconfig)
	assert.NoError(t, err)
	f.Close()
	os.Setenv("KUBECONFIG", f.Name())
	defer os.Setenv("KUBECONFIG", origkc)

	assert.NoError(t, err)
	// defer os.RemoveAll(dir)

	curContext, err := initialContexts(clientcmd.NewDefaultClientConfigLoadingRules())
	assert.NoError(t, err)
	assert.NotNil(t, curContext)

	assert.Equal(t, "foo", curContext)

}
