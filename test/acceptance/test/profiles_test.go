package acceptance

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Weave GitOps Profiles API", func() {
	var (
		namespace        = wego.DefaultNamespace
		clusterName      string
		appRepoRemoteURL string
		tip              TestInputs
		wegoService      = "wego-app"
		wegoPort         = "9001"
		clientSet        *kubernetes.Clientset
		kClient          client.Client
		profileName      = "podinfo"
		resp             []byte
		statusCode       int
	)

	BeforeEach(func() {
		Expect(FileExists(gitopsBinaryPath)).To(BeTrue())
		Expect(githubOrg).NotTo(BeEmpty())

		var err error
		clusterName, _, err = ResetOrCreateCluster(namespace, true)
		Expect(err).NotTo(HaveOccurred())

		tip = generateTestInputs()
		_ = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, true, githubOrg)

		clientSet, kClient = buildKubernetesClients()
	})

	AfterEach(func() {
		cleanupFinalizers(clusterName, namespace)
		deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, githubOrg)
		deleteNamespace(namespace)
	})

	It("@skipOnNightly gets deployed and is accessible via the service", func() {
		By("Installing the Profiles API and setting up the profile helm repository")
		appRepoRemoteURL = "git@github.com:" + githubOrg + "/" + tip.appRepoName + ".git"
		installAndVerifyWego(namespace, appRepoRemoteURL)
		deployProfilesHelmRepository(kClient, namespace)

		By("Getting a list of profiles")
		Eventually(func() int {
			resp, statusCode, err = kubernetesDoRequest(namespace, wegoService, wegoPort, "/v1/profiles", clientSet)
			return statusCode
		}, "60s", "1s").Should(Equal(http.StatusOK))
		Expect(err).NotTo(HaveOccurred())

		profiles := pb.GetProfilesResponse{}
		Expect(json.Unmarshal(resp, &profiles)).To(Succeed())
		Expect(profiles.Profiles).NotTo(BeNil())
		Expect(profiles.Profiles).To(ConsistOf(&pb.Profile{
			Name:        "podinfo",
			Home:        "https://github.com/stefanprodan/podinfo",
			Description: "Podinfo Helm chart for Kubernetes",
			Sources:     []string{"https://github.com/stefanprodan/podinfo"},
			Maintainers: []*pb.Maintainer{
				{
					Name:  "stefanprodan",
					Email: "stefanprodan@users.noreply.github.com",
				},
			},
			//These have to not be nil for the assertion to work. because proto :shrug:
			Keywords:    []string{},
			Annotations: map[string]string{},
		}))

		getProfilesOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("%s --namespace %s get profiles", gitopsBinaryPath, namespace))
		Expect(getProfilesOutput).To(Equal(`NAME	DESCRIPTION	AVAILABLE_VERSIONS
podinfo	Podinfo Helm chart for Kubernetes	6.0.0,6.0.1
`))

		By("Getting the values for a profile")
		resp, statusCode, err = kubernetesDoRequest(namespace, wegoService, wegoPort, "/v1/profiles/podinfo/6.0.1/values", clientSet)
		Expect(err).NotTo(HaveOccurred())
		Expect(statusCode).To(Equal(http.StatusOK))

		profileValues := pb.GetProfileValuesResponse{}
		Expect(json.Unmarshal(resp, &profileValues)).To(Succeed())
		values, err := base64.StdEncoding.DecodeString(profileValues.Values)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(values)).To(ContainSubstring("# Default values for podinfo"))

		cmd := fmt.Sprintf("%s add profile --name %s --version 6.0.1 --namespace %s --cluster %s --config-repo %s --auto-merge", gitopsBinaryPath, profileName, namespace, clusterName, appRepoRemoteURL)
		By(fmt.Sprintf("Adding a profile to a cluster: %s", cmd))
		stdOut, stdErr := runCommandAndReturnStringOutput(cmd)
		Expect(stdErr).To(BeEmpty())
		Expect(stdOut).To(ContainSubstring(
			fmt.Sprintf(`Adding profile:

Name: podinfo
Version: 6.0.1
Cluster: %s
Namespace: %s`, clusterName, namespace)))

		By("Verifying that the profile has been installed on the cluster")
		Eventually(func() int {
			resp, statusCode, err = kubernetesDoRequest(namespace, clusterName+"-"+profileName, "9898", "/healthz", clientSet)
			return statusCode
		}, "120s", "1s").Should(Equal(http.StatusOK))
	})

	When("profiles are installs into a different namespace", func() {
		BeforeEach(func() {
			//encase its left over from a failed run
			deleteNamespace("other")
			namespaceCreatedMsg := runCommandAndReturnSessionOutput("kubectl create ns other")
			Eventually(namespaceCreatedMsg).Should(gbytes.Say("namespace/other created"))
		})

		AfterEach(func() {
			deleteNamespace("other")
		})

		It("@skipOnNightly should not error", func() {
			By("Installing the Profiles API and setting up the profile helm repository")
			appRepoRemoteURL = "git@github.com:" + githubOrg + "/" + tip.appRepoName + ".git"
			installAndVerifyWego(namespace, appRepoRemoteURL)
			By("And installing a helmrepositroy into that namespace")
			deployProfilesHelmRepository(kClient, "other")
			time.Sleep(time.Second * 20)

			By("Getting a list of profiles should still work")
			_, statusCode, err := kubernetesDoRequest(namespace, wegoService, wegoPort, "/v1/profiles", clientSet)
			Expect(err).NotTo(HaveOccurred())
			Expect(statusCode).To(Equal(http.StatusOK))

			By("There should be no errors in the wego app log from the helm cache")
			log := runCommandAndReturnSessionOutput("kubectl logs ")
			Eventually(log).ShouldNot(gbytes.Say("\\\"other\\\" is forbidden"))
		})
	})
})

func cleanupFinalizers(clusterName, namespace string) {
	session := runCommandAndReturnSessionOutput(fmt.Sprintf("%s flux suspend kustomization -n %s %s-system", gitopsBinaryPath, namespace, clusterName))
	Eventually(session, "60s", "1s").Should(gexec.Exit(0))
	session = runCommandAndReturnSessionOutput(fmt.Sprintf("kubectl -n %s delete helmreleases --all --wait", namespace))
	Eventually(session, "60s", "1s").Should(gexec.Exit(0))
	session = runCommandAndReturnSessionOutput(fmt.Sprintf("kubectl -n %s delete helmrepositories --all --wait", namespace))
	Eventually(session, "60s", "1s").Should(gexec.Exit(0))
	session = runCommandAndReturnSessionOutput(fmt.Sprintf("kubectl -n %s delete kustomizations --all --wait", namespace))
	Eventually(session, "60s", "1s").Should(gexec.Exit(0))
	session = runCommandAndReturnSessionOutput(fmt.Sprintf("kubectl -n %s delete gitrepositories --all --wait", namespace))
	Eventually(session, "60s", "1s").Should(gexec.Exit(0))
}

func buildKubernetesClients() (*kubernetes.Clientset, client.Client) {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	Expect(err).NotTo(HaveOccurred())
	clientSet, err := kubernetes.NewForConfig(config)
	Expect(err).NotTo(HaveOccurred())

	_, kClient, err := kube.NewKubeHTTPClientWithConfig(config, "profiles-test")
	Expect(err).NotTo(HaveOccurred())

	return clientSet, kClient
}

func deployProfilesHelmRepository(rawClient client.Client, namespace string) {
	weaveworksChartsHelmRepository := &sourcev1beta1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1beta1.HelmRepositoryKind,
			APIVersion: sourcev1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weaveworks-charts",
			Namespace: namespace,
		},
		Spec: sourcev1beta1.HelmRepositorySpec{
			URL:      "https://weaveworks.github.io/profiles-examples",
			Interval: metav1.Duration{Duration: time.Minute * 10},
		},
	}
	err = rawClient.Create(context.TODO(), weaveworksChartsHelmRepository)
	Expect(err).NotTo(HaveOccurred())
	Eventually(func() error {
		helmRepo := sourcev1beta1.HelmRepository{}
		err := rawClient.Get(context.TODO(), client.ObjectKeyFromObject(weaveworksChartsHelmRepository), &helmRepo)
		if err != nil {
			return err
		}

		readyCondition := meta.FindStatusCondition(helmRepo.Status.Conditions, "Ready")
		if readyCondition != nil && readyCondition.Status == "True" {
			return nil
		}
		return fmt.Errorf("HelmRepository not ready %v", helmRepo.Status)
	}, "10s", "1s").Should(Succeed())
}

func kubernetesDoRequest(namespace, serviceName, servicePort, path string, clientset *kubernetes.Clientset) ([]byte, int, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, 0, err
	}

	responseWrapper := clientset.CoreV1().Services(namespace).ProxyGet("http", serviceName, servicePort, u.String(), nil)

	data, err := responseWrapper.DoRaw(context.TODO())
	if err != nil {
		if se, ok := err.(*errors.StatusError); ok {
			return nil, int(se.Status().Code), nil
		}

		return nil, 0, err
	}

	return data, http.StatusOK, nil
}
