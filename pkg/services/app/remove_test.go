package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys/osysfakes"
	"github.com/weaveworks/weave-gitops/pkg/runner"

	"sigs.k8s.io/yaml"
)

var application wego.Application

var fluxDir string

var createdResources map[string][]string

var goatPaths map[string]bool

func populateAppRepo() (string, error) {
	dir, err := ioutil.TempDir("", "an-app-dir")
	if err != nil {
		return "", err
	}

	workloadPath1 := filepath.Join(dir, "kustomize", "one", "path", "to", "files")
	workloadPath2 := filepath.Join(dir, "kustomize", "another", "path", "to", "more", "files")
	if err := os.MkdirAll(workloadPath1, 0777); err != nil {
		return "", err
	}
	if err := os.MkdirAll(workloadPath2, 0777); err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(filepath.Join(workloadPath1, "nginx.yaml"), []byte("file1"), 0644); err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(filepath.Join(workloadPath2, "nginx.yaml"), []byte("file2"), 0644); err != nil {
		return "", err
	}

	return dir, nil
}

func sliceRemove(item string, slice []string) []string {
	location := 0

	for idx, val := range slice {
		if item == val {
			location = idx
			break
		}
	}
	return append(slice[:location], slice[location+1:]...)
}

func storeCreatedResource(manifestData []byte) error {
	manifests := bytes.Split(manifestData, []byte("\n---\n"))
	for _, manifest := range manifests {
		manifestMap := map[string]interface{}{}

		if err := yaml.Unmarshal(manifest, &manifestMap); err != nil {
			return err
		}

		metamap := manifestMap["metadata"].(map[string]interface{})
		kind := manifestMap["kind"].(string)

		if createdResources[kind] == nil {
			createdResources[kind] = []string{}
		}

		createdResources[kind] = append(createdResources[kind], metamap["name"].(string))
	}
	return nil
}

func storeGOATPath(path string) {
	goatPaths[path] = true
}

func setupFlux() error {
	dir, err := ioutil.TempDir("", "a-home-dir")
	if err != nil {
		return err
	}

	fluxDir = dir
	cliRunner := &runner.CLIRunner{}
	osysClient := &osysfakes.FakeOsys{}
	fluxClient := flux.New(osysClient, cliRunner)
	osysClient.UserHomeDirStub = func() (string, error) {
		return dir, nil
	}
	appSrv.(*App).flux = fluxClient
	fluxBin, err := ioutil.ReadFile(filepath.Join("..", "..", "flux", "bin", "flux"))
	if err != nil {
		return err
	}

	binPath, err := fluxClient.GetBinPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(binPath, 0777)
	if err != nil {
		return err
	}

	exePath, err := fluxClient.GetExePath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(exePath, fluxBin, 0777)
	if err != nil {
		return err
	}

	return nil
}

func runAddAndCollectResources(addParams AddParams) ([]ResourceRef, error) {
	params, err := appSrv.(*App).updateParametersIfNecessary(addParams)
	if err != nil {
		return nil, err
	}
	if err := appSrv.Add(params); err != nil {
		return nil, err
	}
	info := getAppResourceInfo(makeWegoApplication(params), "test-cluster")
	return info.clusterResources(), nil
}

func checkResults(appResources []ResourceRef) error {
	fmt.Printf("CR: %#+v\n", createdResources)
	for _, res := range appResources {
		resources := createdResources[res.kind]
		if len(resources) == 0 {
			return fmt.Errorf("expected resources to be created")
		}
		createdResources[res.kind] = sliceRemove(res.name, resources)
	}

	for kind, leftovers := range createdResources {
		if len(leftovers) > 0 {
			return fmt.Errorf("unexpected %s resources: %#+v\n", kind, leftovers)
		}
	}
	return nil
}

var _ = Describe("Remove", func() {
	var _ = BeforeEach(func() {
		application = makeWegoApplication(AddParams{
			Url:            "https://github.com/foo/bar",
			Path:           "./kustomize",
			Branch:         "main",
			Dir:            ".",
			DeploymentType: "kustomize",
			Namespace:      "wego-system",
			AppConfigUrl:   "NONE",
			AutoMerge:      true,
		})
	})

	It("gives a correct error message when app path not found", func() {
		application.Spec.Path = "./badpath"
		appRepoDir, err := populateAppRepo()
		Expect(err).ShouldNot(HaveOccurred())
		defer os.RemoveAll(appRepoDir)
		_, err = findAppManifests(application, appRepoDir)
		Expect(err).Should(MatchError("application path './badpath' not found"))
	})

	It("locates application manifests", func() {
		appRepoDir, err := populateAppRepo()
		Expect(err).ShouldNot(HaveOccurred())
		defer os.RemoveAll(appRepoDir)
		manifests, err := findAppManifests(application, appRepoDir)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(manifests)).To(Equal(2))
		for _, manifest := range manifests {
			Expect(manifest).To(Or(Equal([]byte("file1")), Equal([]byte("file2"))))
		}
	})

	Context("Collecting resources deployed to cluster", func() {
		var _ = BeforeEach(func() {
			Expect(setupFlux()).To(Succeed())

			gitClient.WriteStub = func(path string, manifest []byte) error {
				storeGOATPath(path)
				return storeCreatedResource(manifest)
			}

			kubeClient.ApplyStub = func(manifest []byte, namespace string) ([]byte, error) {
				if err := storeCreatedResource(manifest); err != nil {
					return nil, err
				}
				return []byte(""), nil
			}
		})

		var _ = AfterEach(func() {
			os.RemoveAll(fluxDir)
		})

		Context("Collecting resources for helm charts", func() {
			var _ = BeforeEach(func() {
				addParams = AddParams{
					Url:            "https://charts.kube-ops.io",
					Branch:         "main",
					DeploymentType: "helm",
					Namespace:      "wego-system",
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				goatPaths = map[string]bool{}
				createdResources = map[string][]string{}
			})

			It("collects cluster resources for helm chart from helm repo with configURL = NONE", func() {
				addParams.Chart = "loki"

				appResources, err := runAddAndCollectResources(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(appResources)).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = NONE", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.Path = "./"

				appResources, err := runAddAndCollectResources(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(appResources)).To(Succeed())
			})

			It("collects cluster resources for helm chart from helm repo with configURL = <url>", func() {
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Chart = "loki"

				appResources, err := runAddAndCollectResources(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(appResources)).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = <url>", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Path = "./"

				appResources, err := runAddAndCollectResources(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(appResources)).To(Succeed())
			})
		})

		Context("Collecting resources for non-helm apps", func() {
			var _ = BeforeEach(func() {
				addParams = AddParams{
					Url:            "ssh://git@github.com/user/wego-fork-test.git",
					Branch:         "main",
					DeploymentType: "kustomize",
					Namespace:      "wego-system",
					Path:           "./",
					AppConfigUrl:   "NONE",
					AutoMerge:      true,
				}

				goatPaths = map[string]bool{}
				createdResources = map[string][]string{}
			})

			It("collects cluster resources for non-helm app with configURL = NONE", func() {
				appResources, err := runAddAndCollectResources(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(appResources)).To(Succeed())
			})

			It("collects cluster resources for non-helm app configURL = ''", func() {
				addParams.AppConfigUrl = ""

				appResources, err := runAddAndCollectResources(addParams)
				Expect(err).ShouldNot(HaveOccurred())
				fmt.Printf("AR: %#+v\n", appResources)
				Expect(checkResults(appResources)).To(Succeed())
			})

			It("collects cluster resources for non-helm app with configURL = <url>", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"

				appResources, err := runAddAndCollectResources(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(appResources)).To(Succeed())
			})
		})
	})
})
