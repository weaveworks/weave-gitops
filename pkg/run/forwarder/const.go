package forwarder

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
)

type portForwardAPodRequest struct {
	RestConfig *rest.Config                // RestConfig is the kubernetes config
	Pod        v1.Pod                      // Pod is the selected pod for this port forwarding
	LocalPort  int                         // LocalPort is the local port that will be selected to expose the PodPort
	PodPort    int                         // PodPort is the target port for the pod
	Streams    genericclioptions.IOStreams // Steams configures where to write or read input from
	StopCh     <-chan struct{}             // StopCh is the channel used to manage the port forward lifecycle
	ReadyCh    chan struct{}               // ReadyCh communicates when the tunnel is ready to receive traffic
}

type carry struct {
	StopCh  chan struct{}              // StopCh is the channel used to manage the port forward lifecycle
	ReadyCh chan struct{}              // ReadyCh communicates when the tunnel is ready to receive traffic
	PF      *portforward.PortForwarder // the instance of Portforwarder
}

type PodOption struct {
	LocalPort int    // the local port for forwarding
	PodPort   int    // the k8s pod port
	Pod       v1.Pod // the k8s pod metadata
}

type Option struct {
	LocalPort   int    // the local port for forwarding
	RemotePort  int    // the remote port port for forwarding
	Namespace   string // the k8s namespace metadata
	PodName     string // the k8s pod metadata
	ServiceName string // the k8s service metadata
	Source      string // the k8s source string, eg: svc/my-nginx-svc po/my-nginx-66b6c48dd5-ttdb2
}

type Result struct {
	Close func()                                        // close the port forwarding
	Ready func() ([][]portforward.ForwardedPort, error) // block till the forwarding ready
	Wait  func()                                        // block and listen IOStreams close signal
}
