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

var createdResources map[string]map[string]bool

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
			createdResources[kind] = map[string]bool{}
		}

		createdResources[kind][metamap["name"].(string)] = true
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

func runAddAndCollectInfo(addParams AddParams) (*AppResourceInfo, error) {
	params, err := appSrv.(*App).updateParametersIfNecessary(addParams)
	if err != nil {
		return nil, err
	}
	if err := appSrv.Add(params); err != nil {
		return nil, err
	}
	return getAppResourceInfo(makeWegoApplication(params), "test-cluster"), nil
}

func checkResults(info *AppResourceInfo) error {
	appResources := info.clusterResources()
	for _, res := range appResources {
		resources := createdResources[res.kind]
		if len(resources) == 0 {
			return fmt.Errorf("expected %s resources to be created", res.kind)
		}
		delete(resources, res.name)
	}

	for kind, leftovers := range createdResources {
		if len(leftovers) > 0 {
			return fmt.Errorf("unexpected %s resources: %#+v\n", kind, leftovers)
		}
	}

	for _, path := range info.clusterResourcePaths() {
		delete(goatPaths, path)
	}

	if len(goatPaths) > 0 {
		return fmt.Errorf("unexpected paths: %#+v\n", goatPaths)
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
				createdResources = map[string]map[string]bool{}
			})

			It("collects cluster resources for helm chart from helm repo with configURL = NONE", func() {
				addParams.Chart = "loki"

				info, err := runAddAndCollectInfo(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(info)).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = NONE", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.Path = "./"

				info, err := runAddAndCollectInfo(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(info)).To(Succeed())
			})

			It("collects cluster resources for helm chart from helm repo with configURL = <url>", func() {
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Chart = "loki"

				info, err := runAddAndCollectInfo(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(info)).To(Succeed())
			})

			It("collects cluster resources for helm chart from git repo with configURL = <url>", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
				addParams.Path = "./"

				info, err := runAddAndCollectInfo(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(info)).To(Succeed())
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
				createdResources = map[string]map[string]bool{}
			})

			It("collects cluster resources for non-helm app with configURL = NONE", func() {
				info, err := runAddAndCollectInfo(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(info)).To(Succeed())
			})

			It("collects cluster resources for non-helm app configURL = ''", func() {
				addParams.AppConfigUrl = ""

				info, err := runAddAndCollectInfo(addParams)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(checkResults(info)).To(Succeed())
			})

			It("collects cluster resources for non-helm app with configURL = <url>", func() {
				addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
				addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"

				info, err := runAddAndCollectInfo(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(checkResults(info)).To(Succeed())
			})
		})
	})
})
