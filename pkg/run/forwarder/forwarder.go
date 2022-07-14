package forwarder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/sync/errgroup"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var once sync.Once

//WithForwarders forward ports per the options. Cancel the context will stop the forwarder.
func WithForwarders(ctx context.Context, config *rest.Config, options []*Option) (*Result, error) {
	return forwarders(ctx, options, config)
}

// It is to forward port for k8s cloud services.
func forwarders(ctx context.Context, options []*Option, config *rest.Config) (*Result, error) {
	newOptions, err := parseOptions(options)
	if err != nil {
		return nil, err
	}

	podOptions, err := handleOptions(ctx, newOptions, config)
	if err != nil {
		return nil, err
	}

	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	carries := make([]*carry, len(podOptions))

	var g errgroup.Group

	for index, option := range podOptions {
		index := index
		stopCh := make(chan struct{}, 1)
		readyCh := make(chan struct{})

		req := &portForwardAPodRequest{
			RestConfig: config,
			Pod:        option.Pod,
			LocalPort:  option.LocalPort,
			PodPort:    option.PodPort,
			Streams:    stream,
			StopCh:     stopCh,
			ReadyCh:    readyCh,
		}

		g.Go(func() error {
			pf, err := portForwardAPod(req)
			if err != nil {
				return err
			}
			carries[index] = &carry{StopCh: stopCh, ReadyCh: readyCh, PF: pf}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	ret := &Result{
		Close: func() {
			once.Do(func() {
				for _, c := range carries {
					close(c.StopCh)
				}
			})
		},
		Ready: func() ([][]portforward.ForwardedPort, error) {
			pfs := [][]portforward.ForwardedPort{}
			for _, c := range carries {
				<-c.ReadyCh
				ports, err := c.PF.GetPorts()
				if err != nil {
					return nil, err
				}
				pfs = append(pfs, ports)
			}
			return pfs, nil
		},
	}

	ret.Wait = func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		fmt.Println("Bye...")
		ret.Close()
	}

	go func() {
		<-ctx.Done()
		ret.Close()
	}()

	return ret, nil
}

// It is to forward port, and return the forwarder.
func portForwardAPod(req *portForwardAPodRequest) (*portforward.PortForwarder, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return nil, err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)

	if err != nil {
		return nil, err
	}

	go func() {
		if err := fw.ForwardPorts(); err != nil {
			panic(err)
		}
	}()

	return fw, nil
}
