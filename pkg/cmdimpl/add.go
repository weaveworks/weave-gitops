package cmdimpl

// Implementation of the 'wego add' command

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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
	Owner          string
	Url            string
	Path           string
	Branch         string
	PrivateKey     string
	DeploymentType string
	Namespace      string
	DryRun         bool
	IsPrivate      bool
}

var (
	params AddParamSet
)

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
	if params.DryRun {
		fmt.Printf(cmd + "\n")
	} else {

		_, err = fluxops.CallFlux(cmd)
		if err != nil {
			return nil, wrapError(err, "could not create git secret")
		}
	}

	cmd = fmt.Sprintf(`create source git "wego" \
        --url="ssh://git@github.com/%s/%s" \
        --branch="%s" \
        --secret-ref="wego" \
        --interval=30s \
        --export \
        --namespace=%s`,
		getOwner(),
		fluxRepoName,
		params.Branch,
		params.Namespace)
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

func generateSourceManifest() ([]byte, error) {
	secretName := params.Name

	cmd := fmt.Sprintf(`create secret git "%s" \
			--url="%s" \
			--private-key-file="%s" \
			--namespace=%s`,
		secretName,
		params.Url,
		params.PrivateKey,
		params.Namespace)
	if params.DryRun {
		fmt.Printf(cmd + "\n")
	} else {

		_, err := fluxops.CallFlux(cmd)

		if err != nil {
			return nil, wrapError(err, "could not create git secret")
		}
	}

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
	sourceManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create git source")
	}
	return sourceManifest, nil
}

func generateKustomizeManifest() ([]byte, error) {
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
	if err != nil {
		return nil, wrapError(err, "could not create kustomization manifest")
	}

	return kustomizeManifest, nil
}

func generateHelmManifest() ([]byte, error) {
	cmd := fmt.Sprintf(`create helmrelease %s \
			--source="GitRepository/%s" \
			--chart="%s" \
			--interval=5m \
			--export \
			--namespace=%s`,
		params.Name,
		params.Name,
		params.Path,
		params.Namespace,
	)
	return fluxops.CallFlux(cmd)
}

func getOwner() string {
	owner, err := fluxops.GetOwnerFromEnv()
	if err != nil || owner == "" {
		owner = getOwnerFromUrl(params.Url)
	}

	// command flag has priority
	if params.Owner != "" {
		return params.Owner
	}

	return owner
}

// ie: ssh://git@github.com/weaveworks/some-repo
func getOwnerFromUrl(url string) string {
	parts := strings.Split(url, "/")

	return parts[len(parts)-2]
}

func commitAndPush(files ...string) error {
	cmdStr := `git pull --rebase && \
                git add %s && \
                git commit -m 'Save %s' && \
                git push`

	if params.DryRun {
		fmt.Fprintf(shims.Stdout(), cmdStr+"\n", strings.Join(files, " "), strings.Join(files, ", "))
		return nil
	}

	cmd := fmt.Sprintf(cmdStr, strings.Join(files, " "), strings.Join(files, ", "))
	output, err := utils.CallCommand(cmd)

	if err != nil {
		return wrapError(err, string(output))
	}
	return nil
}

func wrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// Add provides the implementation for the wego add command
func Add(args []string, allParams AddParamSet) error {
	if len(args) < 1 {
		return errors.New("location of application not specified")
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
		return errors.New("WeGO not installed... exiting\n")
	}

	// Set up wego repository if required
	fluxRepoName, err := fluxops.GetRepoName()
	if err != nil {
		return wrapError(err, "could not get flux repo name")
	}

	reposDir := filepath.Join(os.Getenv("HOME"), ".wego", "repositories")
	fluxRepo := filepath.Join(reposDir, fluxRepoName)
	appSubdir := filepath.Join(fluxRepo, "apps", params.Name)

	if !params.DryRun {
		if err := os.MkdirAll(appSubdir, 0755); err != nil {
			return wrapError(err, "could not make app subdir")
		}
	}

	owner := getOwner()

	if !params.DryRun {
		if err := os.Chdir(fluxRepo); err != nil {
			return wrapError(err, "could not chdir to flux repo")
		}
	}

	if err := utils.CallCommandForEffect(fmt.Sprintf("git ls-remote ssh://git@github.com/%s/%s.git", owner, fluxRepoName)); err != nil {
		fmt.Printf("wego repo does not exist, it will be created...\n")
		if !params.DryRun {
			if err := utils.CallCommandForEffectWithDebug("git init"); err != nil {
				return wrapError(err, "could not init git repo")
			}
		}

		cmdStr := `git remote add origin git@github.com:%s/%s.git && \
            git pull --rebase origin main && \
            git checkout main && \
            git push --set-upstream origin main`
		cmd := fmt.Sprintf(cmdStr, owner, fluxRepoName)

		if err := gitproviders.CreateRepository(fluxRepoName, owner, params.IsPrivate); err != nil {
			return wrapError(err, "could not create repository")
		}

		if !params.DryRun {
			if err := utils.CallCommandForEffectWithDebug(cmd); err != nil {
				return wrapError(err, "could not add remote")
			}
		} else {
			fmt.Fprint(shims.Stdout(), cmd)
		}
	} else if !params.DryRun {
		if err := utils.CallCommandForEffectWithDebug("git branch --set-upstream-to=origin/main main"); err != nil {
			return wrapError(err, "could not set upstream branch")
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

	if !params.DryRun {
		kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)

		if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoSource)); err != nil {
			return wrapError(err, "could not apply wego source")
		}

		if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoKust)); err != nil {
			return wrapError(err, "could not apply wego kustomization")
		}

	} else {
		fmt.Fprintf(shims.Stdout(), "Applying wego platform resources...")
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
	source, err := generateSourceManifest()
	if err != nil {
		return wrapError(err, "could not generate source manifest")
	}

	fmt.Println("DeploymentType check1", params.DeploymentType)
	var appManifests []byte
	switch params.DeploymentType {
	case string(DeployTypeHelm):
		appManifests, err = generateHelmManifest()
	case string(DeployTypeKustomize):
		appManifests, err = generateKustomizeManifest()
	default:
		return fmt.Errorf("deployment type not supported: %s", params.DeploymentType)
	}

	if err != nil {
		return wrapError(err, "error generating manifest")
	}

	sourceName := filepath.Join(appSubdir, "source-"+params.Name+".yaml")
	manifestsName := filepath.Join(appSubdir, fmt.Sprintf("%s-%s.yaml", params.DeploymentType, params.Name))
	appYamlName := filepath.Join(appSubdir, "app.yaml")

	if !params.DryRun {
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
	} else {
		fmt.Fprintf(shims.Stdout(), "Applying wego resources for application...")
	}

	fmt.Printf("Successfully added repository: %s.\n", params.Name)

	return nil
}
