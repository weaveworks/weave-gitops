package cmdimpl

// Implementation of the 'wego add' command

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	cgitprovider "github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

const appYamlTemplate = `apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: {{ .AppName }}
spec:
  path: {{ .AppPath }}
  url: {{ .AppURL }}
`

const (
	DeployTypeKustomize = "kustomize"
	DeployTypeHelm      = "helm"
)

type AddParamSet struct {
	Dir            string
	Name           string
	Url            string
	Path           string
	Branch         string
	PrivateKey     string
	DeploymentType string
	Namespace      string
}

var (
	params    AddParamSet
	repoOwner string
)

// checkError will print a message to stderr and exit
func checkError(msg string, err interface{}) {
	if err != nil {
		fmt.Fprintf(shims.Stderr(), "%s: %v\n", msg, err)
		shims.Exit(1)
	}
}

func checkAddError(err interface{}) {
	checkError("Failed to add workload repository", err)
}

func getClusterRepoName() (string, error) {
	clusterName, err := status.GetClusterName()
	if err != nil {
		return "", wrapError(err, "could not get cluster name")
	}
	return clusterName + "-wego", nil
}

func updateParametersIfNecessary() error {
	if params.Name == "" {
		repoPath, err := filepath.Abs(params.Dir)
		if err != nil {
			return wrapError(err, "could not get directory")
		}

		repoName := strings.ReplaceAll(filepath.Base(repoPath), "_", "-")

		name, err := getClusterRepoName()
		if err != nil {
			return wrapError(err, "could not update parameters")
		}

		params.Name = name + "-" + repoName
	}

	if params.Url == "" {
		urlout, err := exec.Command("git", "remote", "get-url", "origin").CombinedOutput()
		if err != nil {
			return wrapError(err, "could not get remote origin url")
		}

		url := strings.TrimRight(string(urlout), "\n")
		fmt.Printf("URL not specified; ")
		params.Url = url
	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(params.Url, sshPrefix) {
		trimmed := strings.TrimPrefix(params.Url, sshPrefix)
		params.Url = "ssh://git@github.com/" + trimmed
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.Url)

	return nil

}

func generateWegoSourceManifest() ([]byte, error) {
	fluxRepoName, err := fluxops.GetRepoName()
	if err != nil {
		return nil, wrapError(err, "could not get flux repo name")
	}

	cmd := fmt.Sprintf(`create secret git "wego" \
		--url="ssh://git@github.com/%s/%s" \
		--private-key-file="%s" \
		--namespace=%s`,
		getOwner(),
		fluxRepoName,
		params.PrivateKey,
		params.Namespace)
	fmt.Println("debug3", cmd)
	_, err = fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create git secret")
	}

	cmd = fmt.Sprintf(`create source git "wego" \
		--url="ssh://git@github.com/%s/%s" \
		--branch="%s" \
		--secret-ref="wego" \
		--interval=30s \
		--export \
		--namespace=%s `,
		getOwner(),
		fluxRepoName,
		params.Branch,
		params.Namespace)
	fmt.Println("debug2", cmd)
	sourceManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create source")
	}

	return sourceManifest, nil
}

func generateWegoKustomizeManifest() ([]byte, error) {
	cmd := fmt.Sprintf(`create kustomization "wego" \
		--path="./" \
		--source="wego" \
		--prune=true \
		--validation=client \
		--interval=5m \
		--export \
		--namespace=%s`,
		params.Namespace)
	kustomizeManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create kustomization")
	}

	return kustomizeManifest, nil
}

func generateSourceManifest() []byte {
	secretName := params.Name

	cmd := fmt.Sprintf(`create secret git "%s" \
			--url="%s" \
			--private-key-file="%s" \
			--namespace=%s`,
		secretName,
		params.Url,
		params.PrivateKey,
		params.Namespace)
	_, err := fluxops.CallFlux(cmd)
	checkAddError(err)

	cmd = fmt.Sprintf(`create source git "%s" \
			--url="%s" \
			--branch="%s" \
			--secret-ref="%s" \
			--interval=30s \
			--export \
			--namespace=%s `,
		params.Name,
		params.Url,
		params.Branch,
		secretName,
		params.Namespace)
	fmt.Println("debug1", cmd)
	sourceManifest, err := fluxops.CallFlux(cmd)
	checkAddError(err)
	return sourceManifest
}

func generateKustomizeManifest() []byte {

	cmd := fmt.Sprintf(`create kustomization "%s" \
				--path="%s" \
				--source="%s" \
				--prune=true \
				--validation=client \
				--interval=5m \
				--export \
				--namespace=%s`,
		params.Name,
		params.Path,
		params.Name,
		params.Namespace)
	kustomizeManifest, err := fluxops.CallFlux(cmd)
	checkAddError(err)
	return kustomizeManifest
}

func generateHelmManifest() []byte {
	helmManifest, err := fluxops.CallFlux(
		fmt.Sprintf(`create helmrelease %s --source="GitRepository/%s" --chart="%s" --interval=5m --export`, params.Name, params.Name, params.Path))

	checkAddError(err)
	return helmManifest
}

func getOwner() string {
	if repoOwner != "" {
		return repoOwner
	}
	owner, err := fluxops.GetOwnerFromEnv()
	if err != nil || owner == "" {
		repoOwner = getOwnerInteractively()
		return repoOwner
	}
	repoOwner = owner
	return owner
}

func getOwnerInteractively() string {
	fmt.Printf("Who is the owner of the repository? ")
	reader := bufio.NewReader(shims.Stdin())
	str, err := reader.ReadString('\n')
	checkAddError(err)

	if str == "\n" {
		return getOwnerInteractively()
	}

	return strings.Trim(str, "\n")
}

func commitAndPush(files ...string) error {
	cmd := fmt.Sprintf(`git pull --rebase && \
				git add %s && \
				git commit -m 'Save %s' && \
				git push`,
		strings.Join(files, " "),
		strings.Join(files, ", "))
	_, err := utils.CallCommand(cmd)
	return err
}

func wrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// Add provides the implementation for the wego add command
func Add(args []string, allParams AddParamSet) error {
	if len(args) < 1 {
		fmt.Printf("Location of application not specified.\n")
		shims.Exit(1)
	}
	params = allParams
	params.Dir = args[0]
	fmt.Printf("Updating parameters from environment... ")
	if err := updateParametersIfNecessary(); err != nil {
		return wrapError(err, "could not update parameters")
	}
	fmt.Printf("done\n\n")
	fmt.Printf("Checking cluster status... ")
	clusterStatus := status.GetClusterStatus()
	fmt.Printf("%s\n\n", clusterStatus)

	if clusterStatus == status.Unmodified {
		fmt.Printf("WeGO not installed... exiting\n")
		shims.Exit(1)
	}

	// Set up wego repository if required
	fluxRepoName, err := fluxops.GetRepoName()
	if err != nil {
		return wrapError(err, "could not get repo name")
	}

	reposDir := filepath.Join(os.Getenv("HOME"), ".wego", "repositories")
	fluxRepo := filepath.Join(reposDir, fluxRepoName)
	appSubdir := filepath.Join(fluxRepo, "apps", params.Name)
	if err := os.MkdirAll(appSubdir, 0755); err != nil {
		return wrapError(err, "could not make app subdir")
	}

	owner := getOwner()
	if err := os.Chdir(fluxRepo); err != nil {
		return wrapError(err, "could not chdir")
	}

	if err := utils.CallCommandForEffect(fmt.Sprintf("git ls-remote ssh://git@github.com/%s/%s.git", owner, fluxRepoName)); err != nil {
		fmt.Printf("wego repo does not exist, it will be created...\n")

		if err := utils.CallCommandForEffectWithDebug("git init"); err != nil {
			return wrapError(err, "could not do git init")
		}

		c, err := cgitprovider.GithubProvider()
		if err != nil {
			return wrapError(err, "could not get Github provider")
		}

		orgRef := cgitprovider.NewOrgRepositoryRef(cgitprovider.GITHUB_DOMAIN, owner, fluxRepoName)

		repoInfo := cgitprovider.NewRepositoryInfo("wego repo", gitprovider.RepositoryVisibilityPrivate)

		repoCreateOpts := &gitprovider.RepositoryCreateOptions{
			AutoInit:        gitprovider.BoolVar(true),
			LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
		}

		if err := cgitprovider.CreateOrgRepository(c, orgRef, repoInfo, repoCreateOpts); err != nil {
			return wrapError(err, "could not create org repo")
		}

		cmd := fmt.Sprintf(`git remote add origin %s && \
			git pull --rebase origin main && \
			git checkout main && \
			git push --set-upstream origin main`,
			orgRef.String())

		if err := utils.CallCommandForEffectWithDebug(cmd); err != nil {
			return wrapError(err, "could not configure org repo")
		}
	} else {
		cmd := "git branch --set-upstream-to=origin/main main"

		if err := utils.CallCommandForEffectWithDebug(cmd); err != nil {
			return wrapError(err, "could not set upstream")
		}
	}

	// Install Source and Kustomize controllers, and CRD for application (may already be present)
	wegoSource, err := generateWegoSourceManifest()
	if err != nil {
		return wrapError(err, "could not generate wego source manifest")
	}

	wegoKust, err := generateWegoKustomizeManifest()
	if err != nil {
		return wrapError(err, "could not generate wego kustomize manifest")
	}

	kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)

	if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoSource)); err != nil {
		return wrapError(err, "could not apply wego source")
	}

	if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoKust)); err != nil {
		return wrapError(err, "could not apply wego kustomization")
	}

	// Create app.yaml
	t, err := template.New("appYaml").Parse(appYamlTemplate)
	if err != nil {
		return wrapError(err, "could not parse app yaml template")
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		AppName string
		AppPath string
		AppURL  string
	}{params.Name, params.Path, params.Url})
	if err != nil {
		return wrapError(err, "could not execute populated template")
	}

	// Create flux custom resources for new repo being added
	source := generateSourceManifest()

	var appManifests []byte
	if params.DeploymentType == DeployTypeHelm {
		appManifests = generateHelmManifest()
	} else {
		appManifests = generateKustomizeManifest()
	}

	sourceName := filepath.Join(appSubdir, "source-"+params.Name+".yaml")
	manifestsName := filepath.Join(appSubdir, fmt.Sprintf("%s-%s.yaml", params.DeploymentType, params.Name))
	appYamlName := filepath.Join(appSubdir, "app.yaml")

	if err := ioutil.WriteFile(sourceName, source, 0644); err != nil {
		return wrapError(err, "could not write source")
	}

	if err := ioutil.WriteFile(manifestsName, appManifests, 0644); err != nil {
		return wrapError(err, "could not write app manifests")
	}

	if err := ioutil.WriteFile(appYamlName, populated.Bytes(), 0644); err != nil {
		return wrapError(err, "could not write app yaml populated template")
	}

	if err := commitAndPush(sourceName, manifestsName, appYamlName); err != nil {
		return wrapError(err, "could not commit and/or push")
	}

	fmt.Printf("Successfully added repository: %s.\n", params.Name)

	return nil
}
