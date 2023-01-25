package watch

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/mattn/go-tty"
	"github.com/pkg/browser"
	"github.com/weaveworks/weave-gitops/core/logger"
	clilogger "github.com/weaveworks/weave-gitops/pkg/logger"
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

type PortForwardShortcut struct {
	Name     string
	HostPort string
}

var PortForwardShortcutRunes = []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

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

func ForwardPort(log logr.Logger, pod *corev1.Pod, cfg *rest.Config, specMap *PortForwardSpec, waitFwd chan struct{}, readyChannel chan struct{}) error {
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

	outStd := bytes.Buffer{}
	outErr := bytes.Buffer{}

	fw, err2 := portforward.NewOnAddresses(
		dialer,
		[]string{"localhost"},
		[]string{fmt.Sprintf("%s:%s", specMap.HostPort, specMap.ContainerPort)},
		waitFwd,
		readyChannel,
		&outStd,
		&outErr,
	)

	// TODO: these should probably use a separate goroutine and fluxexec.writeOutput,
	// but they won't log much information so this is enough
	if outStd.Len() > 0 {
		log.V(logger.LogLevelInfo).Info(outStd.String())
	}

	if outErr.Len() > 0 {
		log.V(logger.LogLevelError).Info(outErr.String())
	}

	if err2 != nil {
		return err2
	}

	return fw.ForwardPorts()
}

func ShowPortForwards(ctx context.Context, log clilogger.Logger, portForwards map[rune]PortForwardShortcut) {
	// print keyboard shortcuts
	// print text in bold
	fmt.Printf("\n\033[1m%s\033[0m\n\n", "We set up port forwards for you, use the number below to open it in the browser")

	keys := getSortedPortForwardKeys(portForwards)

	for _, key := range keys {
		portForward := portForwards[key]

		fmt.Printf("(%c) %s: http://localhost:%s\n", key, portForward.Name, portForward.HostPort)
	}

	fmt.Println()

	// open tty
	tt, err := tty.Open()
	if err != nil {
		log.Failuref("Error opening tty: %v", err)
		return
	}

	// close tty on exit
	go func(ctx context.Context) {
		<-ctx.Done()

		if err := tt.Close(); err != nil {
			log.Failuref("Error closing tty: %v", err)
		}
	}(ctx)

	// listen for keypresses
	go func() {
		for {
			r, err := tt.ReadRune()
			if err != nil {
				log.Failuref("Error reading keypress: %v", err)
			}

			portForward, ok := portForwards[r]

			if ok {
				err = browser.OpenURL(fmt.Sprintf("http://localhost:%s", portForward.HostPort))
				if err != nil {
					log.Failuref("Error opening portforward URL: %v", err)
				}
			}
		}
	}()
}

func GetNextPortForwardKey(portForwards map[rune]PortForwardShortcut) (rune, error) {
	numPortForwards := len(portForwards)
	numPortForwardShortcuts := len(PortForwardShortcutRunes) - 1

	if numPortForwards > numPortForwardShortcuts-1 {
		return 0, fmt.Errorf("too many port forwards, max is %d", numPortForwardShortcuts)
	}

	return PortForwardShortcutRunes[numPortForwards+1], nil
}

func getSortedPortForwardKeys(portForwards map[rune]PortForwardShortcut) []rune {
	keys := make([]rune, 0)
	for k := range portForwards {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}
