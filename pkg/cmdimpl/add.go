package cmdimpl

// Implementation of the 'wego add' command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type DeploymentType string

const appYamlTemplate = `apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: {{ .AppName }}
spec:
  path: {{ .AppPath }}
  url: {{ .AppURL }}
`

const (
	DeployTypeKustomize DeploymentType = "kustomize"
	DeployTypeHelm      DeploymentType = "helm"
)

type AddParamSet struct {
	Dir            string
	Name           string
	Owner          string
	Url            string
	Path           string
	Branch         string
	PrivateKey     string
	PrivateKeyPass string
	DeploymentType string
	Namespace      string
	DryRun         bool
	IsPrivate      bool
}

var (
	params AddParamSet
)

type AddDependencies struct {
	GitClient git.Git
}

func getClusterRepoName() (string, error) {
	clusterName, err := status.GetClusterName()
	if err != nil {
		return "", wrapError(err, "could not get cluster name")
	}
	return clusterName + "-wego", nil
}

func updateParametersIfNecessary(gitClient git.Git) error {
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
		// gitClient := git.New(nil)
		repo, err := gitClient.Open(params.Dir)
		if err != nil {
			return wrapError(err, fmt.Sprintf("failed to open repository: %s", params.Dir))
		}

		remote, err := repo.Remote("origin")
		if err != nil {
			return err
		}

		urls := remote.Config().URLs

		if len(urls) == 0 {
			return fmt.Errorf("remote config in %s does not have an url", params.Dir)
		}

		params.Url = urls[0]
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

	owner, err := getOwner()
	if err != nil {
		return nil, err
	}

	cmd := fmt.Sprintf(`create secret git "wego" \
        --url="ssh://git@github.com/%s/%s" \
        --private-key-file="%s" \
        --namespace=%s`,
		owner,
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
		owner,
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

func getOwner() (string, error) {
	owner, err := fluxops.GetOwnerFromEnv()
	if err != nil || owner == "" {
		owner, err = getOwnerFromUrl(params.Url)
		if err != nil {
			return "", fmt.Errorf("could not get owner %s", err)
		}
	}

	// command flag has priority
	if params.Owner != "" {
		return params.Owner, nil
	}

	return owner, nil
}

// ie: ssh://git@github.com/weaveworks/some-repo
func getOwnerFromUrl(url string) (string, error) {
	parts := strings.Split(url, "/")

	if len(parts) < 2 {
		return "", fmt.Errorf("could not get owner from url %s", url)
	}

	return parts[len(parts)-2], nil
}

func commitAndPush(ctx context.Context, gitClient git.Git) error {
	_, err := gitClient.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
		Message: "Add App manifests",
	})
	if err != nil && err != git.ErrNoStagedFiles {
		return fmt.Errorf("failed to commit sync manifests: %w", err)
	}
	if err == nil {
		fmt.Println("Pushing app manifests to repository")
		if err = gitClient.Push(ctx); err != nil {
			return fmt.Errorf("failed to push manifests: %w", err)
		}
	} else {
		fmt.Println("App manifests are up to date")
	}

	return nil
}

func wrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// Add provides the implementation for the wego add command
func Add(args []string, allParams AddParamSet, deps *AddDependencies) error {
	ctx := context.Background()

	if len(args) < 1 {
		return errors.New("location of application not specified")
	}

	params = allParams
	params.Dir = args[0]
	fmt.Printf("Updating parameters from environment... ")
	if err := updateParametersIfNecessary(deps.GitClient); err != nil {
		return wrapError(err, "could not update parameters")
	}
	fmt.Printf("done\n\n")
	fmt.Printf("Checking cluster status... ")
	clusterStatus := status.GetClusterStatus()
	fmt.Printf("%s\n\n", clusterStatus)

	if clusterStatus == status.Unmodified {
		return errors.New("WeGO not installed... exiting")
	}

	// Set up wego repository if required
	wegoRepoName, err := fluxops.GetRepoName()
	if err != nil {
		return wrapError(err, "could not get flux repo name")
	}

	owner, err := getOwner()
	if err != nil {
		return err
	}

	fmt.Printf("Verifying %s repository exists...\n", wegoRepoName)
	if _, err := gitproviders.RepositoryExists(wegoRepoName, owner); err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			fmt.Printf("Creating %s repository...\n", wegoRepoName)
			if !params.DryRun {
				if err := gitproviders.CreateRepository(wegoRepoName, owner, params.IsPrivate); err != nil {
					return wrapError(err, "could not create repository")
				}
			}
		} else {
			return wrapError(err, "could not check repository exists")
		}
	}

	wegoRepoDir, err := ioutil.TempDir("", wegoRepoName)
	if err != nil {
		return err
	}
	defer os.RemoveAll(wegoRepoDir)

	gitClient := deps.GitClient

	wegoRepoURL := fmt.Sprintf("ssh://git@github.com/%s/%s.git", owner, wegoRepoName)
	fmt.Printf("Cloning %s...\n", wegoRepoURL)
	if !params.DryRun {
		if _, err := gitClient.Clone(ctx, wegoRepoDir, wegoRepoURL, params.Branch); err != nil {
			return wrapError(err, fmt.Sprintf("could not clone repository: %s", wegoRepoURL))
		}
	}

	fmt.Fprintf(shims.Stdout(), "Applying wego platform resources...\n")
	if !params.DryRun {
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
		source, err := generateSourceManifest()
		if err != nil {
			return wrapError(err, "could not generate source manifest")
		}

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

		appSubdir := filepath.Join("apps", params.Name)
		sourcePath := filepath.Join(appSubdir, "source-"+params.Name+".yaml")
		manifestsPath := filepath.Join(appSubdir, fmt.Sprintf("%s-%s.yaml", params.DeploymentType, params.Name))
		appYamlPath := filepath.Join(appSubdir, "app.yaml")

		if err := gitClient.Write(sourcePath, source); err != nil {
			return wrapError(err, "could not write source")
		}

		if err := gitClient.Write(manifestsPath, appManifests); err != nil {
			return wrapError(err, "could not write app manifests")
		}

		if err := gitClient.Write(appYamlPath, populated.Bytes()); err != nil {
			return wrapError(err, "could not write app yaml populated template")
		}
	}

	fmt.Fprintf(shims.Stdout(), "Commiting and pushing wego resources for application...\n")
	if !params.DryRun {
		if err := commitAndPush(ctx, gitClient); err != nil {
			return wrapError(err, "could not commit and/or push")
		}
	}

	fmt.Printf("Successfully added %s to the repository.\n", params.Name)

	return nil
}
