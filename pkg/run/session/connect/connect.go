package connect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	log "github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run/session/find"
	"github.com/weaveworks/weave-gitops/pkg/run/session/localkubernetes"
	"github.com/weaveworks/weave-gitops/pkg/run/session/portforward"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Connection struct {
	Address                   string
	BackgroundProxy           bool
	Context                   string
	Insecure                  bool
	KubeConfig                string
	KubeConfigContextName     string
	LocalPort                 int
	Log                       log.Logger
	Namespace                 string
	PodName                   string
	Print                     bool
	Server                    string
	ServiceAccount            string
	ServiceAccountClusterRole string
	ServiceAccountExpiration  int
	UpdateCurrent             bool

	errorChan        chan error
	interruptChan    chan struct{}
	kubeClient       *kubernetes.Clientset
	kubeClientConfig clientcmd.ClientConfig
	portForwarding   bool
	rawConfig        api.Config
	restConfig       *rest.Config
}

func (conn *Connection) Connect(vclusterName string, command []string) error {
	if vclusterName == "" {
		return fmt.Errorf("vcluster name is required")
	}

	if conn.PodName == "" {
		return fmt.Errorf("pod name is required")
	}

	// prepare clients and find vcluster
	err := conn.prepare(vclusterName)
	if err != nil {
		return err
	}

	// retrieve vcluster kube config
	kubeConfig, err := conn.getVClusterKubeConfig(vclusterName, command)
	if err != nil {
		return err
	}

	if len(command) == 0 && conn.ServiceAccount == "" && conn.Server == "" && conn.BackgroundProxy && localkubernetes.IsDockerInstalledAndUpAndRunning() {
		// start background container
		server, err := localkubernetes.CreateBackgroundProxyContainer(vclusterName, conn.Namespace, &conn.rawConfig, kubeConfig, conn.LocalPort, conn.Log)
		if err != nil {
			conn.Log.Warningf("Error exposing local vcluster, will fallback to port-forwarding: %v", err)
			conn.BackgroundProxy = false
		}
		conn.Server = server
	}

	// check if we should execute command
	if len(command) > 0 {
		return conn.executeCommand(*kubeConfig, command)
	}

	// write kube config to buffer
	out, err := clientcmd.Write(*kubeConfig)
	if err != nil {
		return err
	}

	// write kube config to file
	if conn.UpdateCurrent {
		var clusterConfig *api.Cluster
		for _, c := range kubeConfig.Clusters {
			clusterConfig = c
		}

		var authConfig *api.AuthInfo
		for _, a := range kubeConfig.AuthInfos {
			authConfig = a
		}

		err = updateKubeConfig(conn.KubeConfigContextName, clusterConfig, authConfig, true)
		if err != nil {
			return err
		}

		conn.Log.Successf("Switched active kube context to %s", conn.KubeConfigContextName)
		if !conn.BackgroundProxy && conn.portForwarding {
			conn.Log.Warningf("Since you are using port-forwarding to connect, you will need to leave this terminal open")
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-c
				kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).RawConfig()
				if err == nil && kubeConfig.CurrentContext == conn.KubeConfigContextName {
					err = deleteContext(&kubeConfig, conn.KubeConfigContextName, conn.Context)
					if err != nil {
						conn.Log.Failuref("Error deleting context: %v", err)
					} else {
						conn.Log.Actionf("Switched back to context %v", conn.Context)
					}
				}
				os.Exit(1)
			}()

			defer func() {
				signal.Stop(c)
			}()
			conn.Log.Println("- Use CTRL+C to return to your previous kube context\n")
			conn.Log.Println("- Use `kubectl get namespaces` in another terminal to access the vcluster\n")
		} else {
			conn.Log.Println("- Use `vcluster disconnect` to return to your previous kube context\n")
			conn.Log.Println("- Use `kubectl get namespaces` to access the vcluster\n")
		}
	} else if conn.Print {
		_, err = os.Stdout.Write(out)
		if err != nil {
			return err
		}
	} else {
		err = os.WriteFile(conn.KubeConfig, out, 0666)
		if err != nil {
			return errors.Wrap(err, "write kube config")
		}

		conn.Log.Successf("Virtual cluster kube config written to: %s", conn.KubeConfig)
		if conn.Server == "" {
			conn.Log.Println(fmt.Sprintf("- Use `vcluster connect %s -n %s -- kubectl get ns` to execute a command directly within this terminal\n", vclusterName, conn.Namespace))
		}
		conn.Log.Println(fmt.Sprintf("- Use `kubectl --kubeconfig %s get namespaces` to access the vcluster\n", conn.KubeConfig))
	}

	// wait for port-forwarding if necessary
	if conn.portForwarding {
		if conn.Server != "" {
			// Stop port-forwarding here
			close(conn.interruptChan)
		}

		return <-conn.errorChan
	}

	return nil
}

func (conn *Connection) executeCommand(vKubeConfig api.Config, command []string) error {
	if !conn.portForwarding {
		return fmt.Errorf("command is specified, but port-forwarding isn't started")
	}

	defer close(conn.interruptChan)

	// wait for vcluster to be ready
	err := conn.waitForVCluster(vKubeConfig, conn.errorChan)
	if err != nil {
		return err
	}

	// convert to local kube config
	vKubeConfig = conn.getLocalVClusterConfig(vKubeConfig)
	out, err := clientcmd.Write(vKubeConfig)
	if err != nil {
		return err
	}

	// write a temporary kube file
	tempFile, err := os.CreateTemp("", "")
	if err != nil {
		return errors.Wrap(err, "create temp file")
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(tempFile.Name())

	_, err = tempFile.Write(out)
	if err != nil {
		return errors.Wrap(err, "write kube config to temp file")
	}

	err = tempFile.Close()
	if err != nil {
		return errors.Wrap(err, "close temp file")
	}

	commandErrChan := make(chan error)
	execCmd := exec.Command(command[0], command[1:]...)
	execCmd.Env = os.Environ()
	execCmd.Env = append(execCmd.Env, "KUBECONFIG="+tempFile.Name())
	execCmd.Stdout = os.Stdout
	execCmd.Stdin = os.Stdin
	execCmd.Stderr = os.Stderr
	err = execCmd.Start()
	if err != nil {
		return err
	}
	go func() {
		commandErrChan <- execCmd.Wait()
	}()

	select {
	case err := <-conn.errorChan:
		if execCmd.Process != nil {
			_ = execCmd.Process.Kill()
		}

		return errors.Wrap(err, "error port-forwarding")
	case err := <-commandErrChan:
		if exitError, ok := err.(*exec.ExitError); ok {
			conn.Log.Failuref("Error executing command: %v", err)
			os.Exit(exitError.ExitCode())
		}

		return err
	}
}

func (conn *Connection) prepare(vclusterName string) error {
	if conn.LocalPort == 0 {
		conn.LocalPort = randomPort()
	}

	if conn.ServiceAccountClusterRole != "" && conn.ServiceAccount == "" {
		return fmt.Errorf("expected service-account to be defined as well")
	}

	var (
		kubeConfigLoader clientcmd.ClientConfig
		vCluster         *find.VCluster
		err              error
	)
	if vclusterName != "" {
		vCluster, err = find.GetVCluster(conn.Context, vclusterName, conn.Namespace)
		if err != nil {
			return err
		}

		kubeConfigLoader = vCluster.ClientFactory
		conn.Context = vCluster.Context
		conn.Namespace = vCluster.Namespace
	} else {
		kubeConfigLoader = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{
			CurrentContext: conn.Context,
		})
	}

	restConfig, err := kubeConfigLoader.ClientConfig()
	if err != nil {
		return errors.Wrap(err, "load kube config")
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return errors.Wrap(err, "create kube client")
	}
	rawConfig, err := kubeConfigLoader.RawConfig()
	if err != nil {
		return errors.Wrap(err, "load raw config")
	}
	rawConfig.CurrentContext = conn.Context

	conn.kubeClient = kubeClient
	conn.restConfig = restConfig
	conn.kubeClientConfig = kubeConfigLoader
	conn.rawConfig = rawConfig

	// set the namespace correctly
	if conn.Namespace == "" {
		conn.Namespace, _, err = kubeConfigLoader.Namespace()
		if err != nil {
			return err
		}
	}

	return nil
}

func (conn *Connection) waitForVCluster(vKubeConfig api.Config, errorChan chan error) error {
	vKubeClient, err := conn.getLocalVClusterClient(vKubeConfig)
	if err != nil {
		return err
	}

	err = wait.PollImmediate(time.Millisecond*200, time.Minute*3, func() (bool, error) {
		select {
		case err := <-errorChan:
			return false, err
		default:
			// check if service account exists
			_, err = vKubeClient.CoreV1().ServiceAccounts("default").Get(context.TODO(), "default", metav1.GetOptions{})
			return err == nil, nil
		}
	})
	if err != nil {
		return errors.Wrap(err, "wait for vcluster to become ready")
	}

	return nil
}

func (conn *Connection) getVClusterKubeConfig(vclusterName string, command []string) (*api.Config, error) {
	var err error
	podName := conn.PodName
	if podName == "" {
		waitErr := wait.PollImmediate(time.Second, time.Second*6, func() (bool, error) {
			// get vcluster pod name
			var pods *corev1.PodList
			pods, err = conn.kubeClient.CoreV1().Pods(conn.Namespace).List(context.Background(), metav1.ListOptions{
				LabelSelector: "app=vcluster,release=" + vclusterName,
			})
			if err != nil {
				return false, err
			} else if len(pods.Items) == 0 {
				err = fmt.Errorf("can't find a running vcluster pod in namespace %s", conn.Namespace)
				return false, nil
			}

			// sort by newest
			sort.Slice(pods.Items, func(i, j int) bool {
				return pods.Items[i].CreationTimestamp.Unix() > pods.Items[j].CreationTimestamp.Unix()
			})
			if pods.Items[0].DeletionTimestamp != nil {
				err = fmt.Errorf("can't find a running vcluster pod in namespace %s", conn.Namespace)
				return false, nil
			}

			podName = pods.Items[0].Name
			return true, nil
		})
		if waitErr != nil {
			return nil, fmt.Errorf("finding vcluster pod: %v - %v", waitErr, err)
		}
	}

	// get the kube config from the Secret
	kubeConfig, err := GetKubeConfig(context.Background(), conn.kubeClient, vclusterName, conn.Namespace, conn.Log)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse kube config")
	}

	// find out port we should listen to locally
	if len(kubeConfig.Clusters) != 1 {
		return nil, fmt.Errorf("unexpected kube config")
	}

	// exchange context name in virtual kube config
	err = conn.exchangeContextName(kubeConfig, vclusterName)
	if err != nil {
		return nil, err
	}

	// check if the vcluster is exposed and set server
	if vclusterName != "" && conn.Server == "" && len(command) == 0 {
		err = conn.setServerIfExposed(vclusterName, kubeConfig)
		if err != nil {
			return nil, err
		}
	}

	// find out vcluster server port
	port := "8443"
	for k := range kubeConfig.Clusters {
		if conn.Insecure {
			kubeConfig.Clusters[k].CertificateAuthorityData = nil
			kubeConfig.Clusters[k].InsecureSkipTLSVerify = true
		}

		if conn.Server != "" {
			if !strings.HasPrefix(conn.Server, "https://") {
				conn.Server = "https://" + conn.Server
			}

			kubeConfig.Clusters[k].Server = conn.Server
		} else {
			parts := strings.Split(kubeConfig.Clusters[k].Server, ":")
			if len(parts) != 3 {
				return nil, fmt.Errorf("unexpected server in kubeconfig: %s", kubeConfig.Clusters[k].Server)
			}

			port = parts[2]
			parts[2] = strconv.Itoa(conn.LocalPort)
			kubeConfig.Clusters[k].Server = strings.Join(parts, ":")
		}
	}

	// start port forwarding
	if conn.ServiceAccount != "" || conn.Server == "" || len(command) > 0 {
		conn.portForwarding = true
		conn.interruptChan = make(chan struct{})
		conn.errorChan = make(chan error)

		// silence port-forwarding if a command is used
		stdout := io.Writer(os.Stdout)
		stderr := io.Writer(os.Stderr)
		if len(command) > 0 || conn.BackgroundProxy {
			stdout = io.Discard
			stderr = io.Discard
		}

		go func() {
			conn.errorChan <- portforward.StartPortForwardingWithRestart(conn.restConfig, conn.Address, podName, conn.Namespace, strconv.Itoa(conn.LocalPort), port, conn.interruptChan, stdout, stderr, conn.Log)
		}()
	}

	// we want to use a service account token in the kube config
	if conn.ServiceAccount != "" {
		token, err := conn.createServiceAccountToken(*kubeConfig)
		if err != nil {
			return nil, err
		}

		// set service account token
		for k := range kubeConfig.AuthInfos {
			kubeConfig.AuthInfos[k] = &api.AuthInfo{
				Token:                token,
				Extensions:           make(map[string]runtime.Object),
				ImpersonateUserExtra: make(map[string][]string),
			}
		}
	}

	return kubeConfig, nil
}

func (conn *Connection) setServerIfExposed(vClusterName string, vClusterConfig *api.Config) error {
	printedWaiting := false
	err := wait.PollImmediate(time.Second*2, time.Minute*5, func() (done bool, err error) {
		service, err := conn.kubeClient.CoreV1().Services(conn.Namespace).Get(context.TODO(), vClusterName, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				return true, nil
			}

			return false, err
		}

		// not a load balancer? Then don't wait
		if service.Spec.Type == corev1.ServiceTypeNodePort {
			server, err := localkubernetes.ExposeLocal(vClusterName, conn.Namespace, &conn.rawConfig, vClusterConfig, service, conn.LocalPort, conn.Log)
			if err != nil {
				conn.Log.Warningf("Error exposing local vcluster, will fallback to port-forwarding: %v", err)
			}

			conn.Server = server
			return true, nil
		} else if service.Spec.Type != corev1.ServiceTypeLoadBalancer {
			return true, nil
		}

		if len(service.Status.LoadBalancer.Ingress) == 0 {
			if !printedWaiting {
				conn.Log.Actionf("Waiting for vcluster LoadBalancer ip...")
				printedWaiting = true
			}

			return false, nil
		}

		if service.Status.LoadBalancer.Ingress[0].Hostname != "" {
			conn.Server = service.Status.LoadBalancer.Ingress[0].Hostname
		} else if service.Status.LoadBalancer.Ingress[0].IP != "" {
			conn.Server = service.Status.LoadBalancer.Ingress[0].IP
		}

		if conn.Server == "" {
			return false, nil
		}

		conn.Log.Actionf("Using vcluster %s load balancer endpoint: %s", vClusterName, conn.Server)
		return true, nil
	})
	if err != nil {
		return errors.Wrap(err, "wait for vcluster")
	}

	return nil
}

func (conn *Connection) exchangeContextName(kubeConfig *api.Config, vclusterName string) error {
	if conn.KubeConfigContextName == "" {
		if vclusterName != "" {
			conn.KubeConfigContextName = find.VClusterContextName(vclusterName, conn.Namespace, conn.rawConfig.CurrentContext)
		} else {
			conn.KubeConfigContextName = find.VClusterContextName(conn.PodName, conn.Namespace, conn.rawConfig.CurrentContext)
		}
	}

	// update cluster
	for k := range kubeConfig.Clusters {
		kubeConfig.Clusters[conn.KubeConfigContextName] = kubeConfig.Clusters[k]
		delete(kubeConfig.Clusters, k)
		break
	}

	// update context
	for k := range kubeConfig.Contexts {
		ctx := kubeConfig.Contexts[k]
		ctx.Cluster = conn.KubeConfigContextName
		ctx.AuthInfo = conn.KubeConfigContextName
		kubeConfig.Contexts[conn.KubeConfigContextName] = ctx
		delete(kubeConfig.Contexts, k)
		break
	}

	// update authInfo
	for k := range kubeConfig.AuthInfos {
		kubeConfig.AuthInfos[conn.KubeConfigContextName] = kubeConfig.AuthInfos[k]
		delete(kubeConfig.AuthInfos, k)
		break
	}

	// update current-context
	kubeConfig.CurrentContext = conn.KubeConfigContextName
	return nil
}

func (conn *Connection) getLocalVClusterConfig(vKubeConfig api.Config) api.Config {
	// wait until we can access the virtual cluster
	vKubeConfig = *vKubeConfig.DeepCopy()
	for k := range vKubeConfig.Clusters {
		vKubeConfig.Clusters[k].Server = "https://localhost:" + strconv.Itoa(conn.LocalPort)
	}
	return vKubeConfig
}

func (conn *Connection) getLocalVClusterClient(vKubeConfig api.Config) (kubernetes.Interface, error) {
	vRestConfig, err := clientcmd.NewDefaultClientConfig(conn.getLocalVClusterConfig(vKubeConfig), &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "create virtual rest config")
	}

	vKubeClient, err := kubernetes.NewForConfig(vRestConfig)
	if err != nil {
		return nil, errors.Wrap(err, "create virtual kube client")
	}

	return vKubeClient, nil
}

func SafeConcatName(name ...string) string {
	fullPath := strings.Join(name, "-")
	if len(fullPath) > 63 {
		digest := sha256.Sum256([]byte(fullPath))
		return strings.ReplaceAll(fullPath[0:52]+"-"+hex.EncodeToString(digest[0:])[0:10], ".-", "-")
	}
	return fullPath
}

func (conn *Connection) createServiceAccountToken(vKubeConfig api.Config) (string, error) {
	vKubeClient, err := conn.getLocalVClusterClient(vKubeConfig)
	if err != nil {
		return "", err
	}

	var (
		serviceAccount          = conn.ServiceAccount
		serviceAccountNamespace = "kube-system"
	)
	if strings.Contains(conn.ServiceAccount, "/") {
		splitted := strings.Split(conn.ServiceAccount, "/")
		if len(splitted) != 2 {
			return "", fmt.Errorf("unexpected service account reference, expected ServiceAccountNamespace/ServiceAccountName")
		}

		serviceAccountNamespace = splitted[0]
		serviceAccount = splitted[1]
	}

	audiences := []string{"https://kubernetes.default.svc.cluster.local", "https://kubernetes.default.svc", "https://kubernetes.default"}
	expirationSeconds := int64(10 * 365 * 24 * 60 * 60)
	if conn.ServiceAccountExpiration > 0 {
		expirationSeconds = int64(conn.ServiceAccountExpiration)
	}
	token := ""
	conn.Log.Actionf("Create service account token for %s/%s", serviceAccountNamespace, serviceAccount)
	err = wait.Poll(time.Second, time.Minute*3, func() (bool, error) {
		// check if namespace exists
		_, err := vKubeClient.CoreV1().Namespaces().Get(context.TODO(), serviceAccountNamespace, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) || kerrors.IsForbidden(err) {
				return false, err
			}

			return false, nil
		}

		// check if service account exists
		_, err = vKubeClient.CoreV1().ServiceAccounts(serviceAccountNamespace).Get(context.TODO(), serviceAccount, metav1.GetOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) {
				if serviceAccount == "default" {
					return false, nil
				}

				if conn.ServiceAccountClusterRole != "" {
					// create service account
					_, err = vKubeClient.CoreV1().ServiceAccounts(serviceAccountNamespace).Create(context.TODO(), &corev1.ServiceAccount{
						ObjectMeta: metav1.ObjectMeta{
							Name:      serviceAccount,
							Namespace: serviceAccountNamespace,
						},
					}, metav1.CreateOptions{})
					if err != nil {
						return false, err
					}

					conn.Log.Successf("Created service account %s/%s", serviceAccountNamespace, serviceAccount)
				} else {
					return false, err
				}
			} else if kerrors.IsForbidden(err) {
				return false, err
			} else {
				return false, nil
			}
		}

		// create service account cluster role binding
		if conn.ServiceAccountClusterRole != "" {
			clusterRoleBindingName := SafeConcatName("vcluster", "sa", serviceAccount, serviceAccountNamespace)
			clusterRoleBinding, err := vKubeClient.RbacV1().ClusterRoleBindings().Get(context.TODO(), clusterRoleBindingName, metav1.GetOptions{})
			if err != nil {
				if kerrors.IsNotFound(err) {
					// create cluster role binding
					_, err = vKubeClient.RbacV1().ClusterRoleBindings().Create(context.TODO(), &rbacv1.ClusterRoleBinding{
						ObjectMeta: metav1.ObjectMeta{
							Name: clusterRoleBindingName,
						},
						RoleRef: rbacv1.RoleRef{
							APIGroup: rbacv1.SchemeGroupVersion.Group,
							Kind:     "ClusterRole",
							Name:     conn.ServiceAccountClusterRole,
						},
						Subjects: []rbacv1.Subject{
							{
								Kind:      "ServiceAccount",
								Name:      serviceAccount,
								Namespace: serviceAccountNamespace,
							},
						},
					}, metav1.CreateOptions{})
					if err != nil {
						return false, err
					}

					conn.Log.Successf("Created cluster role binding for cluster role %s", conn.ServiceAccountClusterRole)
				} else if kerrors.IsForbidden(err) {
					return false, err
				} else {
					return false, nil
				}
			} else {
				// if cluster role differs, recreate it
				if clusterRoleBinding.RoleRef.Name != conn.ServiceAccountClusterRole {
					err = vKubeClient.RbacV1().ClusterRoleBindings().Delete(context.TODO(), clusterRoleBindingName, metav1.DeleteOptions{})
					if err != nil {
						return false, err
					}

					conn.Log.Successf("Recreate cluster role binding for service account")
					// this will recreate the cluster role binding in the next iteration
					return false, nil
				}
			}
		}

		// create service account token
		result, err := vKubeClient.CoreV1().ServiceAccounts(serviceAccountNamespace).CreateToken(context.TODO(), serviceAccount, &authenticationv1.TokenRequest{Spec: authenticationv1.TokenRequestSpec{
			Audiences:         audiences,
			ExpirationSeconds: &expirationSeconds,
		}}, metav1.CreateOptions{})
		if err != nil {
			if kerrors.IsNotFound(err) || kerrors.IsForbidden(err) {
				return false, err
			}

			return false, nil
		}

		token = result.Status.Token
		return true, nil
	})
	if err != nil {
		return "", errors.Wrap(err, "create service account token")
	}

	return token, nil
}
