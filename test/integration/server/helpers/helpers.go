//go:build !unittest
// +build !unittest

package helpers

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	glAPI "github.com/xanzy/go-gitlab"

	"github.com/fluxcd/go-git-providers/gitprovider"
	kustomizev2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ghAPI "github.com/google/go-github/v32/github"

	"github.com/fluxcd/source-controller/pkg/sourceignore"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

var ErrPathMismatch = errors.New("path mismatch")
var ErrFileMismatch = errors.New("file mismatch")

type WeGODirectoryFS map[string]interface{}

//go:embed yaml/deployment.yaml
var deploymentYaml []byte

//go:embed yaml/kustomization.yaml
var kustomizationYaml []byte

func CreateRepo(ctx context.Context, gp gitprovider.Client, url string) (gitprovider.OrgRepository, *gitprovider.OrgRepositoryRef, error) {
	ref, err := gitprovider.ParseOrgRepositoryURL(url)
	if err != nil {
		return nil, ref, fmt.Errorf("error parsing url: %w", err)
	}

	defaultBranch := "main"
	repo, _, err := gp.OrgRepositories().Reconcile(ctx, *ref, gitprovider.RepositoryInfo{
		Description:   gitprovider.StringVar("Integration test repo"),
		Visibility:    gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate),
		DefaultBranch: gitprovider.StringVar(defaultBranch),
	}, &gitprovider.RepositoryCreateOptions{AutoInit: gitprovider.BoolVar(true)})

	if err != nil {
		return nil, nil, fmt.Errorf("could not reconcile org repo: %w", err)
	}

	err = utils.WaitUntil(bytes.NewBuffer([]byte{}), 3*time.Second, 9*time.Second, func() error {
		r, err := gp.OrgRepositories().Get(ctx, *ref)
		if err != nil {
			return err
		}

		commits, err := r.Commits().ListPage(ctx, defaultBranch, 1, 0)
		if err != nil {
			return err
		}

		if len(commits) == 0 {
			return fmt.Errorf("there are no commits yet")
		}

		return nil
	})

	return repo, ref, err
}

func addFiles(ctx context.Context, message string, repo gitprovider.OrgRepository, files []gitprovider.CommitFile) error {
	_, err := repo.Commits().Create(ctx, "main", "Initial commit", []gitprovider.CommitFile{
		{
			Path:    gitprovider.StringVar("k8s/deployment.yaml"),
			Content: gitprovider.StringVar(string(deploymentYaml)),
		},
		{
			Path:    gitprovider.StringVar("k8s/kustomization.yaml"),
			Content: gitprovider.StringVar(string(kustomizationYaml)),
		},
	})

	if err != nil {
		return fmt.Errorf("adding files: %w", err)
	}

	return repo.Update(ctx)
}

func NewGithubClient(ctx context.Context, token string) *ghAPI.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return ghAPI.NewClient(tc)
}

const InAppRoot = ".weave-gitops"
const ExternalConfigRoot = ".weave-gitops"

func wegoPath(root, p string) string {
	return filepath.Join(root, p)
}

func appYamlPath(root, name string) string {
	return appPath(root, name, "app.yaml")
}

func automationPath(root, appName string) string {
	return appPath(root, appName, fmt.Sprintf("%s-gitops-deploy.yaml", appName))
}

func sourcePath(root, appName string) string {
	return appPath(root, appName, fmt.Sprintf("%s-gitops-source.yaml", appName))
}

func appKustPath(root, appName string) string {
	return appPath(root, appName, "kustomization.yaml")
}

func userKustPath(root, clusterName string) string {
	return filepath.Join(root, "clusters", clusterName, "user", "kustomization.yaml")
}

func appPath(root, appName, filename string) string {
	return wegoPath(root, filepath.Join("apps", appName, filename))
}

func MakeWeGOFS(root, appName, clusterName string) WeGODirectoryFS {
	return map[string]interface{}{
		appYamlPath(root, appName): &wego.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       wego.ApplicationKind,
				APIVersion: wego.GroupVersion.String(),
			},
		},
		appKustPath(root, appName):      &types.Kustomization{},
		automationPath(root, appName):   &kustomizev2.Kustomization{},
		sourcePath(root, appName):       &sourcev1.GitRepository{},
		userKustPath(root, clusterName): &types.Kustomization{},
	}
}

func GenerateExpectedFS(req *pb.AddApplicationRequest, root, clusterName string, app wego.ApplicationSpec, k kustomizev2.KustomizationSpec, s sourcev1.GitRepositorySpec) WeGODirectoryFS {
	expected := map[string]interface{}{
		appYamlPath(root, req.Name): &wego.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       wego.ApplicationKind,
				APIVersion: wego.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Spec: app,
		},
		automationPath(root, req.Name): &kustomizev2.Kustomization{
			TypeMeta: metav1.TypeMeta{
				Kind:       kustomizev2.KustomizationKind,
				APIVersion: kustomizev2.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Spec: k,
		},
		sourcePath(root, req.Name): &sourcev1.GitRepository{
			TypeMeta: metav1.TypeMeta{
				Kind:       sourcev1.GitRepositoryKind,
				APIVersion: sourcev1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Spec: s,
		},
		appKustPath(root, req.Name): &types.Kustomization{
			TypeMeta: types.TypeMeta{
				Kind:       types.KustomizationKind,
				APIVersion: types.KustomizationVersion,
			},
			MetaData: &types.ObjectMeta{
				Name:      req.Name,
				Namespace: req.Namespace,
			},
			Resources: []string{
				"app.yaml",
				fmt.Sprintf("%s-gitops-deploy.yaml", req.Name),
				fmt.Sprintf("%s-gitops-source.yaml", req.Name),
			},
		},
		userKustPath(root, clusterName): &types.Kustomization{
			TypeMeta: types.TypeMeta{
				Kind:       types.KustomizationKind,
				APIVersion: types.KustomizationVersion,
			},
			MetaData: &types.ObjectMeta{
				Name:      clusterName,
				Namespace: req.Namespace,
			},
			Resources: []string{"../../../apps/" + req.Name},
		},
	}

	return expected
}

func Filenames(fs WeGODirectoryFS) []string {
	keys := []string{}
	for k := range fs {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func GetGithubFilesContents(ctx context.Context, gh *ghAPI.Client, org, repoName string, fs WeGODirectoryFS, files []*ghAPI.CommitFile) (WeGODirectoryFS, error) {
	changes := map[string][]byte{}

	for _, file := range files {
		if *file.Status == "removed" {
			delete(fs, *file.Filename)
			continue
		}

		path := *file.Filename

		b, _, err := gh.Git.GetBlobRaw(ctx, org, repoName, *file.SHA)
		if err != nil {
			return nil, fmt.Errorf("error getting blob for %q: %w", path, err)
		}

		changes[path] = b
	}

	return toK8sObjects(changes, fs)
}

func GetGitlabFilesContents(gl *glAPI.Client, fullRepoPath string, fs WeGODirectoryFS, commitSHA string, files []*glAPI.Diff) (WeGODirectoryFS, error) {
	changes := map[string][]byte{}

	for _, file := range files {
		path := file.OldPath
		if file.DeletedFile {
			delete(fs, path)
			continue
		}

		b, _, err := gl.RepositoryFiles.GetRawFile(fullRepoPath, path, &glAPI.GetRawFileOptions{
			Ref: &commitSHA,
		})
		if err != nil {
			return nil, fmt.Errorf("error getting blob for %q: %w", path, err)
		}

		changes[path] = b
	}

	return toK8sObjects(changes, fs)
}

func toK8sObjects(changes map[string][]byte, fs WeGODirectoryFS) (WeGODirectoryFS, error) {
	for path, change := range changes {
		obj, ok := fs[path]

		if !ok {
			fs[path] = nil
			continue
		}

		if err := yaml.Unmarshal(change, obj); err != nil {
			return nil, fmt.Errorf("error unmarshalling change yaml: %w", err)
		}

		fs[path] = obj
	}

	return fs, nil
}

func CreatePopulatedSourceRepo(ctx context.Context, gp gitprovider.Client, url string) (gitprovider.OrgRepository, *gitprovider.OrgRepositoryRef, error) {
	sourceRepo, sourceRef, err := CreateRepo(ctx, gp, url)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create repo: %w", err)
	}

	if err := addFiles(ctx, "Initial Commit", sourceRepo, []gitprovider.CommitFile{
		{
			Path:    gitprovider.StringVar("k8s/deployment.yaml"),
			Content: gitprovider.StringVar(string(deploymentYaml)),
		},
		{
			Path:    gitprovider.StringVar("k8s/kustomization.yaml"),
			Content: gitprovider.StringVar(string(kustomizationYaml)),
		},
	}); err != nil {
		return nil, nil, fmt.Errorf("could not add files to source repo: %w", err)
	}

	return sourceRepo, sourceRef, nil
}

func DiffFS(actual WeGODirectoryFS, expected WeGODirectoryFS) (string, error) {
	actualFiles := Filenames(actual)
	expectedFiles := Filenames(expected)

	pathDiff := cmp.Diff(actualFiles, expectedFiles)

	if pathDiff != "" {
		return pathDiff, fmt.Errorf("%w: paths mismatch (-actual +expected): %s\n", ErrPathMismatch, pathDiff)
	}

	opt := cmpopts.IgnoreFields(wego.Application{}, "ObjectMeta.Labels")

	for path, expected := range expected {
		result := actual[path]

		diff := cmp.Diff(result, expected, opt)
		if diff != "" {
			return diff, fmt.Errorf("%w: filename %q (-actual +expected):\n%s", ErrFileMismatch, path, diff)
		}
	}

	return "", nil
}

func GetIgnoreSpec() *string {
	ignores := []string{".weave-gitops/"}

	for _, ignore := range []string{sourceignore.ExcludeVCS, sourceignore.ExcludeExt, sourceignore.ExcludeCI, sourceignore.ExcludeExtra} {
		ignores = append(ignores, strings.Split(ignore, ",")...)
	}

	ignoreSpec := strings.Join(ignores, "\n")

	return &ignoreSpec
}
