package watch

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardSpec struct {
	Namespace     string
	Name          string
	Kind          string
	HostPort      string
	ContainerPort string
	Map           map[string]string
}

// parse port forward specin the key-value format of "port=8000:8080,resource=svc/app,namespace=default"
func ParsePortForwardSpec(spec string) (*PortForwardSpec, error) {
	specMap := PortForwardSpec{
		Map: make(map[string]string),
	}
	specMap.Namespace = "default"

	for _, pair := range strings.Split(spec, ",") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid port forward spec: %s", spec)
		}

		if kv[0] == "port" {
			// split into port and host port
			portAndHostPort := strings.Split(kv[1], ":")
			specMap.HostPort = portAndHostPort[0]
			specMap.ContainerPort = portAndHostPort[1]
		} else if kv[0] == "resource" {
			// specMap["resource"] = kv[1]
			// split kv[1] into kind and name
			kindAndName := strings.Split(kv[1], "/")
			if len(kindAndName) != 2 {
				return nil, fmt.Errorf("invalid resource: %s", kv[1])
			}
			specMap.Kind = generalizeKind(kindAndName[0])
			specMap.Name = kindAndName[1]
		} else if kv[0] == "namespace" {
			specMap.Namespace = kv[1]
		} else {
			specMap.Map[kv[0]] = kv[1]
		}
	}

	return &specMap, nil
}

func generalizeKind(kind string) string {
	// switch over kind
	switch kind {
	// if it is po, pod, pods return "pod"
	case "po", "pod", "pods":
		return "pod"
	// if it is svc, service, services return "service"
	case "svc", "service", "services":
		return "service"
	// if it is deployment, deployments return "deployment"
	case "deployment", "deployments":
		return "deployment"
	default:
		return kind
	}
}

func ForwardPort(pod *corev1.Pod, cfg *rest.Config, specMap *PortForwardSpec, waitFwd chan struct{}, readyChannel chan struct{}) error {
	reqURL, err := url.Parse(
		fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward",
			cfg.Host,
			pod.Namespace,
			pod.Name,
		),
	)
	if err != nil {
		return err
	}

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", reqURL)
	fw, err2 := portforward.NewOnAddresses(
		dialer,
		[]string{"localhost"},
		[]string{fmt.Sprintf("%s:%s", specMap.HostPort, specMap.ContainerPort)},
		waitFwd,
		readyChannel,
		os.Stdout,
		os.Stderr,
	)

	if err2 != nil {
		return err2
	}

	return fw.ForwardPorts()
}
