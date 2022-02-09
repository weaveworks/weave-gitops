//go:build !unittest
// +build !unittest

package helpers

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	glAPI "github.com/xanzy/go-gitlab"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluxcd/go-git-providers/gitprovider"
	ghAPI "github.com/google/go-github/v32/github"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/fluxcd/source-controller/pkg/sourceignore"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var ErrPathMismatch = errors.New("path mismatch")
var ErrFileMismatch = errors.New("file mismatch")

type WeGODirectoryFS map[string]interface{}

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

func SetWegoConfig(k8sClient client.Client, namespace string, configRepo string) error {
	ctx := context.Background()

	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kube.WegoConfigMapName,
			Namespace: namespace,
		},
	}

	key := client.ObjectKeyFromObject(cm)
	if err := k8sClient.Get(ctx, key, cm); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		if err := k8sClient.Create(ctx, cm); err != nil {
			return err
		}
	}

	cm.Data = map[string]string{
		"config": fmt.Sprintf(`
WegoNamespace: %s
FluxNamespace: %s
ConfigRepo: %s`, namespace, namespace, configRepo),
	}

	return k8sClient.Update(ctx, cm)
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

func GetIgnoreSpec() *string {
	ignores := []string{".weave-gitops/"}

	for _, ignore := range []string{sourceignore.ExcludeVCS, sourceignore.ExcludeExt, sourceignore.ExcludeCI, sourceignore.ExcludeExtra} {
		ignores = append(ignores, strings.Split(ignore, ",")...)
	}

	ignoreSpec := strings.Join(ignores, "\n")

	return &ignoreSpec
}
