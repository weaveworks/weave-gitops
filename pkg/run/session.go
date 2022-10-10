package run

import (
	"os"
	"os/signal"
	"syscall"

	vcluster "github.com/loft-sh/vcluster/cmd/vclusterctl/cmd"
	"github.com/loft-sh/vcluster/cmd/vclusterctl/flags"
	"github.com/loft-sh/vcluster/cmd/vclusterctl/log"
	"github.com/mitchellh/go-ps"
)

type Session struct {
	name string
}

func (s *Session) Connect(namespace string) error {
	subProcArgs := append(os.Args, "--disable-session", "--allow-k8s-context="+s.name)

	connect := vcluster.ConnectCmd{
		GlobalFlags: &flags.GlobalFlags{
			Silent:    true,
			Namespace: namespace,
		},
		Log:                   log.GetInstance(),
		BackgroundProxy:       false, // must be false to avoid creating the docker container
		KubeConfigContextName: s.name,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c

		thisProc := os.Getpid()
		allProcesses, err := ps.Processes()

		if err != nil {
			return
		}

		for _, proc := range allProcesses {
			if proc.PPid() == thisProc {
				// ok it's a child process, obtain the process object
				procObject, err := os.FindProcess(proc.Pid())
				if err != nil {
					continue
				}

				// and notify it
				if err := procObject.Signal(syscall.SIGUSR1); err != nil {
					return
				}
			}
		}
	}()

	err := connect.Connect(s.name, subProcArgs)

	return err
}

func NewSession(name string) (*Session, error) {
	return &Session{name: name}, nil
}
