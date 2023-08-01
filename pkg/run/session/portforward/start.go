package portforward

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/weaveworks/weave-gitops/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"
)

func StartPortForwardingWithRestart(config *rest.Config, address, pod, namespace, localPort, remotePort string, interrupt chan struct{}, stdout, stderr io.Writer, log log.Logger) error {
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// restart port forwarding
	stopChan, err := StartPortForwarding(config, kubeClient, address, pod, namespace, localPort, remotePort, stdout, stderr, log)
	if err != nil {
		return fmt.Errorf("error starting port forwarding: %v", err)
	}

	for {
		select {
		case <-interrupt:
			close(stopChan)
			return nil
		case <-stopChan:
			log.Actionf("Restarting port forwarding")

			// wait for loft pod to start
			//nolint:staticcheck // deprecated, tracking issue: https://github.com/weaveworks/weave-gitops/issues/3812
			err := wait.PollImmediate(time.Second, time.Minute*10, func() (done bool, err error) {
				pod, err := kubeClient.CoreV1().Pods(namespace).Get(context.Background(), pod, metav1.GetOptions{})
				if err != nil {
					return false, nil
				}
				for _, c := range pod.Status.Conditions {
					if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
						return true, nil
					}
				}
				return false, nil
			})
			if err != nil {
				log.Warningf("error waiting for ready vcluster pod: %v", err)
				continue
			}

			// restart port forwarding
			stopChan, err = StartPortForwarding(config, kubeClient, address, pod, namespace, localPort, remotePort, stdout, stderr, log)
			if err != nil {
				log.Warningf("error starting port forwarding: %v", err)
				continue
			}

			log.Successf("Successfully restarted port forwarding")
		}
	}
}

func StartPortForwarding(config *rest.Config, client kubernetes.Interface, address, pod, namespace, localPort, remotePort string, stdout, stderr io.Writer, log log.Logger) (chan struct{}, error) {
	execRequest := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(namespace).
		SubResource("portforward")

	t, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	if address == "" {
		address = "localhost"
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: t}, "POST", execRequest.URL())
	errChan := make(chan error)
	readyChan := make(chan struct{})
	stopChan := make(chan struct{})
	forwarder, err := NewOnAddresses(dialer, []string{address}, []string{localPort + ":" + remotePort}, stopChan, readyChan, errChan, stdout, stderr)
	if err != nil {
		return nil, err
	}

	go func() {
		err := forwarder.ForwardPorts()
		if err != nil {
			errChan <- err
		}
	}()

	// wait till ready
	select {
	case err = <-errChan:
		return nil, err
	case <-readyChan:
	case <-stopChan:
		return nil, fmt.Errorf("stopped before ready")
	}

	// start watcher
	go func() {
		for {
			select {
			case <-stopChan:
				return
			case err = <-errChan:
				log.Actionf("error during port forwarder: %v", err)
				close(stopChan)
				return
			}
		}
	}()

	return stopChan, nil
}
