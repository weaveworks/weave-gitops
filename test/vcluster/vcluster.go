package vcluster

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"

	_ "embed"
)

//go:embed manifests/vcluster-values.yaml.tpl
var vclusterValues string

//go:embed manifests/vcluster-ingress.yaml.tpl
var vclusterIngress string

//go:embed manifests/nginx-ingress-deploy.yaml
var nginxIngressManifests string

type Factory interface {
	Create(ctx context.Context, name string) (client.Client, error)
	Delete(ctx context.Context, name string) error
}

type factory struct {
	hostClient client.Client
}

func NewFactory() (Factory, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, err
	}

	return &factory{
		hostClient: c,
	}, nil
}

func (c *factory) Create(ctx context.Context, name string) (client.Client, error) {
	namespaceObj := &corev1.Namespace{}
	namespaceObj.Name = name

	if err := c.hostClient.Create(ctx, namespaceObj); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("failed creating namespace %s: %w", name, err)
		}
	}

	if err := createCluster(name); err != nil {
		return nil, fmt.Errorf("failed creating cluster: %w", err)
	}

	configPath, err := connectCluster(name)
	if err != nil {
		return nil, fmt.Errorf("failed connecting cluster: %w", err)
	}

	kubeClientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&clientcmd.ClientConfigLoadingRules{ExplicitPath: configPath}, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed getting vcluster client config: %w", err)
	}

	kubeClientConfig.Timeout = 500 * time.Millisecond

	var vclusterClient client.Client

	err = wait.Poll(time.Second, 5*time.Second, func() (bool, error) {
		fmt.Println("creating cluster client...")
		vclusterClient, err = client.New(kubeClientConfig, client.Options{Scheme: kube.CreateScheme()})
		if err != nil {
			return false, nil
		}

		return true, nil
	})

	return vclusterClient, err
}

func (c *factory) Delete(ctx context.Context, name string) error {
	return nil
}

func createCluster(name string) error {
	if err := createClusterIngress(name); err != nil {
		return fmt.Errorf("failed creating cluster ingress: %w", err)
	}

	if err := appendClusterToEtcHosts(name); err != nil {
		return fmt.Errorf("failed appending cluster to /etc/hosts file: %w", err)
	}

	filename, err := writeVclusterValuesToDisk(name)
	if err != nil {
		return err
	}

	args := []string{
		"create", name,
		"-n", name,
		"-f", filename,
		"--upgrade",
	}

	output, err := exec.Command("vcluster", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error executing vcluster %s: %s", strings.Join(args, " "), string(output))
	}

	return nil
}

func connectCluster(name string) (string, error) {
	vKubeconfigFile, err := ioutil.TempFile(os.TempDir(), "vcluster_e2e_kubeconfig_")
	if err != nil {
		return "", fmt.Errorf("could not create a temporary file: %v", err)
	}

	args := []string{
		"connect", name,
		"-n", name,
		"--kube-config", vKubeconfigFile.Name(),
		"--server", fmt.Sprintf("https://%s.k3s", name),
	}

	output, err := exec.Command("vcluster", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error executing vcluster %s: %s", strings.Join(args, " "), string(output))
	}

	return vKubeconfigFile.Name(), nil
}

func writeVclusterValuesToDisk(name string) (string, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "vcluster-values-")
	if err != nil {
		return "", fmt.Errorf("Cannot create temporary file: %w", err)
	}

	values, err := executeTemplate(vclusterValues, name)
	if err != nil {
		return "", err
	}

	// Example writing to the file
	if _, err = tmpFile.Write(values); err != nil {
		return "", fmt.Errorf("Failed to write to temporary file: %w", err)
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func appendClusterToEtcHosts(name string) error {
	clusterEntry := fmt.Sprintf("172.10.0.150 %s.k3s\n", name)

	present, err := checkClusterIsPresent(clusterEntry)
	if err != nil {
		return err
	}

	if present {
		return nil
	}

	f, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(clusterEntry); err != nil {
		return err
	}

	return nil
}

func checkClusterIsPresent(entry string) (bool, error) {
	b, err := ioutil.ReadFile("/etc/hosts")
	if err != nil {
		return false, err
	}

	isExist, err := regexp.Match(entry, b)
	if err != nil {
		return false, err
	}

	return isExist, nil
}

func executeTemplate(tplData string, clusterName string) ([]byte, error) {
	template, err := template.New(clusterName).Parse(tplData)
	if err != nil {
		return nil, fmt.Errorf("error parsing template %s: %w", clusterName, err)
	}

	yaml := &bytes.Buffer{}

	err = template.Execute(yaml, map[string]string{
		"Name": clusterName,
	})
	if err != nil {
		return nil, fmt.Errorf("error injecting values to template: %w", err)
	}

	return yaml.Bytes(), nil
}

func createClusterIngress(name string) error {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "vcluster-ingress-")
	if err != nil {
		return fmt.Errorf("Cannot create temporary file: %w", err)
	}

	values, err := executeTemplate(vclusterIngress, name)
	if err != nil {
		return err
	}

	// Example writing to the file
	if _, err = tmpFile.Write(values); err != nil {
		return fmt.Errorf("Failed to write to temporary file: %w", err)
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		return err
	}

	args := []string{
		"apply",
		"-f", tmpFile.Name(),
	}

	output, err := exec.Command("kubectl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error applying ingress manifests with kubectl %s: %s", strings.Join(args, " "), string(output))
	}

	return nil
}

func InstallNginxIngressController() error {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "vcluster-ingress-")
	if err != nil {
		return fmt.Errorf("Cannot create temporary file: %w", err)
	}

	if _, err = tmpFile.Write([]byte(nginxIngressManifests)); err != nil {
		return fmt.Errorf("Failed to write to temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	args := []string{
		"apply",
		"-f", tmpFile.Name(),
	}

	output, err := exec.Command("kubectl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error applying ingress manifests with kubectl %s: %s", strings.Join(args, " "), string(output))
	}

	waitCmd := `while [[ $(kubectl get pods -n ingress-nginx -l app.kubernetes.io/component=controller -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for pod" && sleep 1; done`

	output, err = exec.Command("bash", "-c", waitCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error waiting ingress controller to be ready. output=%s error=%s", string(output), err)
	}

	return nil
}

func UpdateHostKubeconfig() error {
	kubeconfig := os.Getenv("KUBECONFIG")

	input, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return err
	}

	output := bytes.Replace(input, []byte("127.0.0.1"), []byte("k3s"), -1)

	return ioutil.WriteFile(kubeconfig, output, 0666)
}

func WaitClusterConnectivity() error {
	waitCmd := `while ! kubectl version; do echo "waiting for cluster connectivity" && sleep 1; done`

	output, err := exec.Command("bash", "-c", waitCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error waiting cluster to be ready. output=%s error=%s", string(output), err)
	}

	return nil
}
