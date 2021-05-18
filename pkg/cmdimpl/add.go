package cmdimpl

// Implementation of the 'wego add' command

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/weave-gitops/pkg/yaml"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	cgitprovider "github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

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

func getClusterRepoName() string {
	clusterName, err := utils.GetClusterName()
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
	fluxRepoName, err := utils.GetWegoRepoName()
	checkAddError(err)

	cmd := fmt.Sprintf(`create secret git "wego" \
		--url="ssh://git@github.com/%s/%s" \
		--private-key-file="%s" \
		--namespace=%s`,
		getOwner(),
		fluxRepoName,
		params.PrivateKey,
		params.Namespace)
	_, err = fluxops.CallFlux(cmd)
	checkAddError(err)

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
			--export`,
		params.Name,
		params.Name,
		params.Path,
	)
	helmManifest, err := fluxops.CallFlux(cmd)
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

	wegoRepoName, err := utils.GetWegoRepoName()
	checkAddError(err)
	fmt.Println("wegoRepoName", wegoRepoName)

	fluxRepo, err := utils.GetWegoLocalPath()
	checkAddError(err)
	fmt.Println("fluxRepo", fluxRepo)

	owner := getOwner()
	checkAddError(os.Chdir(fluxRepo))

	if err := utils.CallCommandForEffect(fmt.Sprintf("git ls-remote ssh://git@github.com/%s/%s.git", owner, wegoRepoName)); err != nil {
		fmt.Printf("wego repo does not exist, it will be created...\n")
		checkAddError(utils.CallCommandForEffectWithDebug("git init"))

		c, err := cgitprovider.GithubProvider()
		checkAddError(err)

		orgRef := cgitprovider.NewOrgRepositoryRef(cgitprovider.GITHUB_DOMAIN, owner, wegoRepoName)

		repoInfo := cgitprovider.NewRepositoryInfo("wego repo", gitprovider.RepositoryVisibilityPrivate)

		repoCreateOpts := &gitprovider.RepositoryCreateOptions{
			AutoInit:        gitprovider.BoolVar(true),
			LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
		}

		err = cgitprovider.CreateOrgRepository(c, orgRef, repoInfo, repoCreateOpts)
		checkAddError(err)

		cmd := fmt.Sprintf(`git remote add origin %s && \
			git pull --rebase origin main && \
			git checkout main && \
			git push --set-upstream origin main`,
			orgRef.String())
		checkAddError(utils.CallCommandForEffectWithDebug(cmd))
	} else {
		cmd := "git branch --set-upstream-to=origin/main main"
		checkAddError(utils.CallCommandForEffectWithDebug(cmd))
	}

	// Install Source and Kustomize controllers, and CRD for application (may already be present)
	wegoSource := generateWegoSourceManifest()
	wegoKust := generateWegoKustomizeManifest()
	kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)
	checkAddError(utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoSource)))
	checkAddError(utils.CallCommandForEffectWithInputPipe(kubectlApply, string(wegoKust)))

	// Create app.yaml
	// TODO refactor AppManager @josecordaz
	yamlManager := yaml.NewAppManager(params.Name)
	err = yamlManager.AddApp(yaml.NewApp(params.Name, args[0], params.Url))
	checkAddError(err)

	// Create flux custom resources for new repo being added
	source := generateSourceManifest()

	var appManifests []byte
	switch params.DeploymentType {
	case DeployTypeHelm:
		appManifests = generateHelmManifest()
	case DeployTypeKustomize:
		appManifests = generateKustomizeManifest()
	default:
		log.Fatalf("deployment type does not supported [%s]", params.DeploymentType)
		os.Exit(1)
	}

	checkAddError(utils.CallCommandForEffectWithInputPipe(kubectlApply, string(source)))
	checkAddError(utils.CallCommandForEffectWithInputPipe(kubectlApply, string(appManifests)))

	wegoAppsPath, err := utils.GetWegoAppPath(params.Name)
	checkAddError(err)
	sourceYamlPath := filepath.Join(wegoAppsPath, "source-"+params.Name+".yaml")
	manifestDeployTypeNamePath := filepath.Join(wegoAppsPath, fmt.Sprintf("%s-%s.yaml", params.DeploymentType, params.Name))
	appYamlPath, err := yaml.GetAppsYamlPath(params.Name)
	checkAddError(err)
	checkAddError(ioutil.WriteFile(sourceYamlPath, source, 0644))
	checkAddError(ioutil.WriteFile(manifestDeployTypeNamePath, appManifests, 0644))

	err = commitAndPush(sourceYamlPath, manifestDeployTypeNamePath, appYamlPath)
	checkAddError(err)

	fmt.Printf("Successfully added repository: %s.\n", params.Name)
}
