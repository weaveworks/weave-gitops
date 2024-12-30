package terraform

import (
	"bytes"
	ctx "context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/run"
)

var kubeConfigArgs *genericclioptions.ConfigFlags

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "terraform",
		Aliases: []string{"tf"},
		Args:    cobra.ExactArgs(1),
		Short:   "Get the runner logs of a Terraform object",
		Example: `
# Get the runner logs of a Terraform object in the "flux-system" namespace
gitops logs terraform --namespace flux-system my-resource
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}

			context, err := cmd.Flags().GetString("context")
			if err != nil {
				return err
			}

			kubeConfigArgs.Namespace = &namespace
			kubeConfigArgs.Context = &context

			cfg, err := kubeConfigArgs.ToRESTConfig()
			if err != nil {
				return err
			}

			// Runner pod is named after the Terraform object
			podName := fmt.Sprintf("%s-tf-runner", args[0])

			clientset, err := kubernetes.NewForConfig(cfg)
			if err != nil {
				return errors.New("error in getting access to K8S")
			}

			pod, err := clientset.CoreV1().Pods(namespace).Get(ctx.Background(), podName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
			podLogs, err := req.Stream(ctx.Background())
			if err != nil {
				return errors.New("error in opening stream")
			}

			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, podLogs)
			if err != nil {
				return errors.New("error in copy information from podLogs to buf")
			}

			fmt.Print(buf.String())

			return nil
		},
	}

	kubeConfigArgs = run.GetKubeConfigArgs()
	kubeConfigArgs.AddFlags(cmd.Flags())
	kubeConfigArgs.KubeConfig = &opts.Kubeconfig

	return cmd
}
