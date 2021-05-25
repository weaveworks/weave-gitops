package cmdimpl

// Implementation of the 'wego add' command

import (
	"bytes"
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

func getClusterRepoName() string {
	clusterName, err := status.GetClusterName()
	checkAddError(err)
	return clusterName + "-wego"
}

func updateParametersIfNecessary() {
	if params.Name == "" {
		repoPath, err := filepath.Abs(params.Dir)
		checkAddError(err)
		repoName := strings.ReplaceAll(filepath.Base(repoPath), "_", "-")
		params.Name = getClusterRepoName() + "-" + repoName
	}

	if params.Url == "" {
		urlout, err := exec.Command("git", "remote", "get-url", "origin").CombinedOutput()
		checkError("Failed to discover URL of remote repository", err)
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
}

func generateWegoSourceManifest() []byte {
	fluxRepoName, err := fluxops.GetRepoName()
	checkAddError(err)

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
		checkAddError(err)
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
	checkAddError(err)
	return sourceManifest
}

func generateWegoKustomizeManifest() []byte {
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
	checkAddError(err)
	return kustomizeManifest
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
	if params.DryRun {
		fmt.Printf(cmd + "\n")
	} else {
		_, err := fluxops.CallFlux(cmd)
		checkAddError(err)
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
	helmManifest, err := fluxops.CallFlux(cmd)
	checkAddError(err)
	return helmManifest
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

func commitAndPush(files ...string) {
	cmdStr := `git pull --rebase && \
                git add %s && \
                git commit -m 'Save %s' && \
                git push`

	if params.DryRun {
		fmt.Fprintf(shims.Stdout(), cmdStr+"\n", strings.Join(files, " "), strings.Join(files, ", "))
		return
	}

	cmd := fmt.Sprintf(cmdStr, strings.Join(files, " "), strings.Join(files, ", "))
	_, err := utils.CallCommand(cmd)
	checkAddError(err)
}

// Add provides the implementation for the wego add command
func Add(args []string, allParams AddParamSet) {
	if len(args) < 1 {
		fmt.Printf("Location of application not specified.\n")
		shims.Exit(1)
	}
	params = allParams
	params.Dir = args[0]
	fmt.Printf("Updating parameters from environment... ")
	updateParametersIfNecessary()
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
	checkAddError(err)

	reposDir := filepath.Join(os.Getenv("HOME"), ".wego", "repositories")
	fluxRepo := filepath.Join(reposDir, fluxRepoName)
	appSubdir := filepath.Join(fluxRepo, "apps", params.Name)

	if !params.DryRun {
		checkAddError(os.MkdirAll(appSubdir, 0755))
	}

	owner := getOwner()

	if !params.DryRun {
		checkAddError(os.Chdir(fluxRepo))
	}

	if err := utils.CallCommandForEffect(fmt.Sprintf("git ls-remote ssh://git@github.com/%s/%s.git", owner, fluxRepoName)); err != nil {
		fmt.Printf("wego repo does not exist, it will be created...\n")
		if !params.DryRun {
			checkAddError(utils.CallCommandForEffectWithDebug("git init"))
		}

		cmdStr := `git remote add origin git@github.com:%s/%s.git && \
            git pull --rebase origin main && \
            git checkout main && \
            git push --set-upstream origin main`
		cmd := fmt.Sprintf(cmdStr, owner, fluxRepoName)

		checkAddError(gitproviders.CreateRepository(fluxRepoName, owner, params.IsPrivate))

		if !params.DryRun {
			checkAddError(utils.CallCommandForEffectWithDebug(cmd))
		} else {
			fmt.Fprint(shims.Stdout(), cmd)
		}
	} else if !params.DryRun {
		checkAddError(utils.CallCommandForEffectWithDebug("git branch --set-upstream-to=origin/main main"))
	}

	// Install Source and Kustomize controllers, and CRD for application (may already be present)
	wegoSource := generateWegoSourceManifest()
	wegoKust := generateWegoKustomizeManifest()
	if !params.DryRun {
		kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)
		checkAddError(utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoSource)))
		checkAddError(utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoKust)))
	} else {
		fmt.Fprintf(shims.Stdout(), "Applying wego platform resources...")
	}

	// Create app.yaml
	t, err := template.New("appYaml").Parse(appYamlTemplate)
	checkAddError(err)

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		AppName string
		AppPath string
		AppURL  string
	}{params.Name, params.Path, params.Url})
	checkAddError(err)

	// Create flux custom resources for new repo being added
	source := generateSourceManifest()

	fmt.Println("DeploymentType check1", params.DeploymentType)
	var appManifests []byte
	switch params.DeploymentType {
	case string(DeployTypeHelm):
		appManifests = generateHelmManifest()
	case string(DeployTypeKustomize):
		appManifests = generateKustomizeManifest()
	default:
		checkAddError(fmt.Errorf("deployment type not supported [%s]", params.DeploymentType))
	}

	sourceName := filepath.Join(appSubdir, "source-"+params.Name+".yaml")
	manifestsName := filepath.Join(appSubdir, fmt.Sprintf("%s-%s.yaml", params.DeploymentType, params.Name))
	appYamlName := filepath.Join(appSubdir, "app.yaml")

	if !params.DryRun {
		checkAddError(ioutil.WriteFile(sourceName, source, 0644))
		checkAddError(ioutil.WriteFile(manifestsName, appManifests, 0644))
		checkAddError(ioutil.WriteFile(appYamlName, populated.Bytes(), 0644))

		commitAndPush(sourceName, manifestsName, appYamlName)
	} else {
		fmt.Fprintf(shims.Stdout(), "Applying wego resources for application...")
	}

	fmt.Printf("Successfully added repository: %s.\n", params.Name)
}
