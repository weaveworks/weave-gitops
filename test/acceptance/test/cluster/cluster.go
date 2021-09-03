package cluster

import (
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type Cluster struct {
	Name           string
	Context        string
	KubeConfigPath string
}

func NewCluster(name string, context string, kubeConfigPath string) *Cluster {
	return &Cluster{
		Name:           name,
		Context:        context,
		KubeConfigPath: kubeConfigPath,
	}
}

func (c *Cluster) Delete() {
	cmd := fmt.Sprintf("kind delete cluster --name %s --kubeconfig %s", c.Name, c.KubeConfigPath)
	command := exec.Command("sh", "-c", cmd)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func (c *Cluster) DeleteKubeConfigFile() {
	err := os.RemoveAll(c.KubeConfigPath)
	Expect(err).ShouldNot(HaveOccurred())
}

type ClusterPool struct {
	cluster        chan *Cluster
	lastCluster    *Cluster
	end            bool
	err            error
	kubeConfigRoot string
}

// TODO: Start generating unit tests for ClusterPool
// TODO: Remove last kubeconfigfile and last cluster after error or on end
// TODO: Hability to pass in the name of the cluster you want
// TODO: Generalize paths of kubeconfig, etc.

func NewClusterPool() *ClusterPool {
	return &ClusterPool{cluster: make(chan *Cluster)}
}

func (c *ClusterPool) GetNextCluster() *Cluster {
	return <-c.cluster
}

func (c *ClusterPool) Generate() error {

	kubeConfigRoot, err := ioutil.TempDir("", "kube-config")
	if err != nil {
		return err
	}
	c.kubeConfigRoot = kubeConfigRoot
	fmt.Println("Creating kube config files on ", kubeConfigRoot)

	go func() {
		for !c.end {
			cluster, err := CreateKindCluster(kubeConfigRoot)
			if err != nil {
				c.err = err
				break
			}
			c.lastCluster = cluster
			c.cluster <- cluster
		}
	}()

	return nil
}

func (c *ClusterPool) Error() error {
	return c.err
}

func (c *ClusterPool) End() {
	c.lastCluster.Delete()
	c.lastCluster.DeleteKubeConfigFile()
	c.end = true
}

func CreateFakeCluster(ind int64) (string, error) {
	return fmt.Sprintf("fakeClusterName%d", ind), nil
}

func CreateKindCluster(rootKubeConfigFilesPath string) (*Cluster, error) {
	supportedProviders := "kind"
	supportedK8SVersions := "1.19.1, 1.20.2, 1.21.1"

	provider, found := os.LookupEnv("CLUSTER_PROVIDER")
	if !found {
		provider = "kind"
	}

	k8sVersion, found := os.LookupEnv("K8S_VERSION")
	if !found {
		k8sVersion = "1.20.2"
	}

	if !strings.Contains(supportedProviders, provider) {
		log.Errorf("Cluster provider %s is not supported for testing", provider)
		return nil, errors.New("Unsupported provider")
	}

	if !strings.Contains(supportedK8SVersions, k8sVersion) {
		log.Errorf("Kubernetes version %s is not supported for testing", k8sVersion)
		return nil, errors.New("Unsupported kubernetes version")
	}

	var cluster *Cluster

	if provider == "kind" {
		clusterName := RandString(6)
		kubeConfigFile := "kube-config-" + clusterName
		kubeConfigPath := filepath.Join(rootKubeConfigFilesPath, kubeConfigFile)
		log.Infof("Creating a kind cluster %s", clusterName)
		err := runCommandPassThrough([]string{}, "./scripts/kind-cluster.sh", clusterName, kubeConfigPath, "kindest/node:v"+k8sVersion)
		if err != nil {
			log.Infof("Failed to create kind cluster")
			log.Fatal(err)
			return nil, err
		}
		cluster = NewCluster(clusterName, "kind-"+clusterName, kubeConfigPath)

	}

	return cluster, nil
}

// Run a command, passing through stdout/stderr to the OS standard streams
func runCommandPassThrough(env []string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if len(env) > 0 {
		cmd.Env = env
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandString(length int) string {
	return StringWithCharset(length, charset)
}
