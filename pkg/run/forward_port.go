package run

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func GetPodFromSpecMap(specMap *PortForwardSpec, kubeClient client.Client) (*corev1.Pod, error) {
	namespacedName := types.NamespacedName{Name: specMap.Name, Namespace: specMap.Namespace}

	switch specMap.Kind {
	case "pod":
		pod := &corev1.Pod{}
		if err := kubeClient.Get(context.Background(), namespacedName, pod); err != nil {
			return nil, err
		}

		return pod, nil
	case "service":
		svc := &corev1.Service{}
		if err := kubeClient.Get(context.Background(), namespacedName, svc); err != nil {
			return nil, fmt.Errorf("error getting service: %s, namespaced Name: %v", err, namespacedName)
		}

		// list pods of the service "svc" by selector in a specific namespace using the controller-runtime client
		podList := &corev1.PodList{}
		if err := kubeClient.List(context.Background(), podList,
			client.MatchingLabelsSelector{
				Selector: labels.Set(svc.Spec.Selector).AsSelector(),
			},
			client.InNamespace(svc.Namespace),
		); err != nil {
			return nil, err
		}

		if len(podList.Items) == 0 {
			return nil, ErrNoPodsForService
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				return &pod, nil
			}
		}

		return nil, ErrNoRunningPodsForService
	case "deployment":
		deployment := &appsv1.Deployment{}
		if err := kubeClient.Get(context.Background(), namespacedName, deployment); err != nil {
			return nil, fmt.Errorf("error getting deployment: %s, namespaced Name: %v", err, namespacedName)
		}

		// list pods of the deployment "deployment" by selector in a specific namespace using the controller-runtime client
		podList := &corev1.PodList{}
		if err := kubeClient.List(context.Background(), podList,
			client.MatchingLabelsSelector{
				Selector: labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector(),
			},
			client.InNamespace(deployment.Namespace),
		); err != nil {
			return nil, err
		}

		if len(podList.Items) == 0 {
			return nil, ErrNoPodsForDeployment
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				return &pod, nil
			}
		}

		return nil, ErrNoRunningPodsForDeployment
	}

	return nil, errors.New("unsupported spec kind")
}
