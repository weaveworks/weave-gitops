package app

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

var workloadYaml1 = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: my-nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: my-nginx
  labels:
    name: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      name: nginx
  template:
    metadata:
      namespace: my-nginx
      labels:
        name: nginx
    spec:
      containers:
      - name: nginx
        image: nginx
        ports:
        - containerPort: 80
`)

var workloadYaml2 = []byte(`apiVersion: v1
kind: Namespace
metadata:
  name: my-nginx2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx2
  namespace: my-nginx2
  labels:
    name: nginx2
spec:
  replicas: 3
  selector:
    matchLabels:
      name: nginx2
  template:
    metadata:
      namespace: my-nginx2
      labels:
        name: nginx2
    spec:
      containers:
      - name: nginx2
        image: nginx
        ports:
        - containerPort: 80
`)

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

	if err := ioutil.WriteFile(filepath.Join(workloadPath1, "nginx.yaml"), []byte(workloadYaml1), 0644); err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(filepath.Join(workloadPath2, "nginx.yaml"), []byte(workloadYaml2), 0644); err != nil {
		return "", err
	}

	return dir, nil
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
			Expect(manifest).To(Or(Equal(workloadYaml1), Equal(workloadYaml2)))
		}
	})
})
