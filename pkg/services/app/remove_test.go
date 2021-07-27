package app

import (
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

var fluxDir string

var createdResources map[string][]string

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
			addParams = AddParams{
				Url:            "https://charts.kube-ops.io",
				Branch:         "main",
				DeploymentType: "helm",
				Namespace:      "wego-system",
				AppConfigUrl:   "NONE",
				AutoMerge:      true,
			}
			dir, err := ioutil.TempDir("", "a-home-dir")
			Expect(err).ShouldNot(HaveOccurred())

			fluxDir = dir
			cliRunner := &runner.CLIRunner{}
			osysClient := &osysfakes.FakeOsys{}
			fluxClient := flux.New(osysClient, cliRunner)
			osysClient.UserHomeDirStub = func() (string, error) {
				return dir, nil
			}
			appSrv.(*App).flux = fluxClient
			fluxBin, err := ioutil.ReadFile(filepath.Join("..", "..", "flux", "bin", "flux"))
			Expect(err).ShouldNot(HaveOccurred())
			binPath, err := fluxClient.GetBinPath()
			Expect(err).ShouldNot(HaveOccurred())
			err = os.MkdirAll(binPath, 0777)
			Expect(err).ShouldNot(HaveOccurred())
			exePath, err := fluxClient.GetExePath()
			Expect(err).ShouldNot(HaveOccurred())
			err = ioutil.WriteFile(exePath, fluxBin, 0777)
			Expect(err).ShouldNot(HaveOccurred())

			createdResources = map[string][]string{}

			kubeClient.ApplyStub = func(manifest []byte, namespace string) ([]byte, error) {
				manifestMap := map[string]interface{}{}

				if err := yaml.Unmarshal(manifest, &manifestMap); err != nil {
					return nil, err
				}

				metamap := manifestMap["metadata"].(map[string]interface{})
				kind := manifestMap["kind"].(string)

				if createdResources[kind] == nil {
					createdResources[kind] = []string{}
				}

				createdResources[kind] = append(createdResources[kind], metamap["name"].(string))
				return []byte(""), nil
			}
		})

		var _ = AfterEach(func() {
			os.RemoveAll(fluxDir)
		})

		It("collects cluster resources for helm with configURL = NONE", func() {
			addParams.Chart = "loki"
			addParams, err := appSrv.(*App).updateParametersIfNecessary(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			err = appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			info := getAppResourceInfo(makeWegoApplication(addParams), "test-cluster")
			appResources := info.clusterResources()

			for _, res := range appResources {
				resources := createdResources[res.kind]
				Expect(resources).To(Not(BeEmpty()))
				createdResources[res.kind] = sliceRemove(res.name, resources)
			}

			for _, leftovers := range createdResources {
				Expect(leftovers).To(BeEmpty())
			}
		})

		It("collects cluster resources for helm with configURL = <url>", func() {
			addParams.Url = "ssh://git@github.com/user/wego-fork-test.git"
			addParams.AppConfigUrl = "ssh://git@github.com/user/external.git"
			addParams.Path = "./"
			addParams, err := appSrv.(*App).updateParametersIfNecessary(addParams)
			Expect(err).ShouldNot(HaveOccurred())
			fmt.Printf("AP: %#+v\n", addParams)
			err = appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			info := getAppResourceInfo(makeWegoApplication(addParams), "test-cluster")
			appResources := info.clusterResources()
			fmt.Printf("CR: %#+v\n", createdResources)
			for _, res := range appResources {
				fmt.Printf("KIND: %s\n", res.kind)
				resources := createdResources[res.kind]
				Expect(resources).To(Not(BeEmpty()))
				createdResources[res.kind] = sliceRemove(res.name, resources)
			}
		})
	})
})
