package names

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders"
)

const MaxKubernetesResourceNameLength = 63

func CreateClusterSourceName(gitSourceURL gitproviders.RepoURL) string {
	provider := string(gitSourceURL.Provider())
	cleanRepoName := replaceUnderscores(gitSourceURL.RepositoryName())
	qualifiedName := fmt.Sprintf("wego-auto-%s-%s", provider, cleanRepoName)
	lengthConstrainedName := hashNameIfTooLong(qualifiedName)

	return lengthConstrainedName
}

func replaceUnderscores(str string) string {
	return strings.ReplaceAll(str, "_", "-")
}

type GeneratedSecretName string

func (s GeneratedSecretName) String() string {
	return string(s)
}

func CreateRepoSecretName(gitSourceURL gitproviders.RepoURL) GeneratedSecretName {
	provider := string(gitSourceURL.Provider())
	cleanRepoName := replaceUnderscores(gitSourceURL.RepositoryName())
	qualifiedName := fmt.Sprintf("wego-%s-%s", provider, cleanRepoName)
	lengthConstrainedName := hashNameIfTooLong(qualifiedName)

	return GeneratedSecretName(lengthConstrainedName)
}

func hashNameIfTooLong(name string) string {
	if !ApplicationNameTooLong(name) {
		return name
	}

	return fmt.Sprintf("wego-%x", md5.Sum([]byte(name)))
}

func ApplicationNameTooLong(name string) bool {
	return len(name) > MaxKubernetesResourceNameLength
}
