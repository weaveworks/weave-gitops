package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
	crenvtest "sigs.k8s.io/controller-runtime/pkg/envtest"
)

const (
	// DefaultImage which will be used if non is provided.
	DefaultImage = "docker.io/rancher/k3s:v1.22.9-rc4-k3s1"
	// InsecurePort assigned to the K3s cluster.
	InsecurePort = "6443"
	// BindAddress assigned to the K3s cluster.
	BindAddress = "0.0.0.0"
	// K3sConfigFileName is the name of the k3s config file.
	K3sConfigFileName = "k3s-test-kubeconfig.yaml"
)

// Environment which will back the Kubebuilder testsuite.
type Environment struct {
	// Image which will be used for spinning up K3s.
	Image string
	// CRDDirectoryPaths for preloading CustomResourceDefinitions.
	// This is field name was taken from the controller-runtime envtest package
	// for compatibility reasons.
	CRDDirectoryPaths []string
	// Internal container identifier.
	id string
}

func getCallerDir() string {
	_, file, _, _ := runtime.Caller(3)

	base := filepath.Base(file)

	return strings.TrimRight(file, base)
}

// Start the test environment.
func (e *Environment) Start() (*rest.Config, error) {
	ctx := context.Background()

	if e.Image == "" {
		e.Image = DefaultImage
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	_, err = cli.ImagePull(ctx, e.Image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}

	natPort, err := nat.NewPort("tcp", InsecurePort)
	if err != nil {
		return nil, err
	}

	k3sOutFilesDir := getCallerDir()

	containerConfig := &container.Config{
		Image: e.Image,
		Cmd: []string{
			"server",
			"--disable=traefik", "--disable=coredns", "--disable=servicelb", "--disable=local-storage", "--disable=metrics-server",
			"--disable-scheduler", "--disable-cloud-controller", "--disable-kube-proxy", "--disable-network-policy",
			"-o", "/output/" + K3sConfigFileName,
		},
		ExposedPorts: nat.PortSet{
			natPort: {},
		},
		Volumes: map[string]struct{}{
			fmt.Sprintf("%s:/output/", k3sOutFilesDir): {},
		},
	}

	containerHostConfig := &container.HostConfig{
		Privileged: true,
		PortBindings: map[nat.Port][]nat.PortBinding{
			natPort: {
				{
					HostIP: BindAddress,
				},
			},
		},
		Mounts: []mount.Mount{
			{Source: k3sOutFilesDir, Target: "/output", Type: mount.TypeBind},
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, containerHostConfig, nil, nil, "")
	if err != nil {
		return nil, err
	}

	e.id = resp.ID

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}

	var config *rest.Config

	if err := wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		inspect, err := cli.ContainerInspect(ctx, resp.ID)
		if err != nil {
			return false, err
		}

		port, err := getContainerPort(inspect.NetworkSettings.Ports, natPort)
		if err != nil {
			fmt.Println(err)
			return false, nil
		}

		config, err = clusterConfigFromPath(filepath.Join(k3sOutFilesDir, K3sConfigFileName))
		if err != nil {
			fmt.Println(err)
			return false, nil
		}

		config.Host = fmt.Sprintf("https://localhost:%s", port)

		kc, err := kclient.New(config, kclient.Options{})
		if err != nil {
			fmt.Println(err)
			return false, nil
		}

		crds := apiextensionsv1.CustomResourceDefinitionList{}
		if err := kc.List(ctx, &crds); err != nil {
			fmt.Println(err)
			return false, nil
		}

		return true, nil
	}); err != nil {
		return nil, err
	}

	_, err = crenvtest.InstallCRDs(config, envtest.CRDInstallOptions{
		Paths: e.CRDDirectoryPaths,
		// Scheme:             kube.CreateScheme(),
		ErrorIfPathMissing: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed installing crds: %w", err)
	}

	return config, nil
}

// Stop the test environment.
func (e *Environment) Stop() error {
	ctx := context.Background()

	_ = os.Remove(filepath.Join(getCallerDir(), K3sConfigFileName))

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	fmt.Println("Removing k3s test container:", e.id)

	err = cli.ContainerStop(ctx, e.id, nil)
	if err != nil {
		return err
	}

	return cli.ContainerRemove(ctx, e.id, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
}

func clusterConfigFromPath(path string) (*rest.Config, error) {
	cfgLoadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: path,
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(cfgLoadingRules, nil).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not create rest config: %w", err)
	}

	return config, nil
}
