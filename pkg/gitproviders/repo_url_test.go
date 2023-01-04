package gitproviders

import (
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = DescribeTable("detectGitProviderFromURL", func(input string, expected GitProviderName) {
	result, err := detectGitProviderFromURL(input, map[string]string{})
	Expect(err).NotTo(HaveOccurred())
	Expect(result).To(Equal(expected))
},
	Entry("ssh+github", "ssh://git@github.com/weaveworks/weave-gitops.git", GitProviderGitHub),
	Entry("ssh+gitlab", "ssh://git@gitlab.com/weaveworks/weave-gitops.git", GitProviderGitLab),
)

var _ = Describe("get owner from url", func() {
	DescribeTable("getOwnerFromURL", func(normalizedURL string, providerName GitProviderName, expected string) {
		u, err := url.Parse(normalizedURL)
		Expect(err).NotTo(HaveOccurred())
		result, err := getOwnerFromURL(*u, providerName)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(expected))
	},
		Entry("github", "ssh://git@github.com/weaveworks/weave-gitops.git", GitProviderGitHub, "weaveworks"),
		Entry("gitlab", "ssh://git@gitlab.com/weaveworks/weave-gitops.git", GitProviderGitLab, "weaveworks"),
		Entry("gitlab with subgroup", "ssh://git@gitlab.com/weaveworks/sub_group/weave-gitops.git", GitProviderGitLab, "weaveworks/sub_group"),
	)

	It("missing owner", func() {
		normalizedURL := "ssh://git@gitlab.com/weave-gitops.git"
		u, err := url.Parse(normalizedURL)
		Expect(err).NotTo(HaveOccurred())
		_, err = getOwnerFromURL(*u, GitProviderGitLab)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("could not get owner from url ssh://git@gitlab.com/weave-gitops.git"))
	})

	It("empty url", func() {
		normalizedURL := ""
		u, err := url.Parse(normalizedURL)
		Expect(err).NotTo(HaveOccurred())
		_, err = getOwnerFromURL(*u, GitProviderGitLab)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("could not get owner from url "))
	})

	It("subgroup in a subgroup", func() {
		normalizedURL := "ssh://git@gitlab.com/weaveworks/sub_group/another_sub_group/weave-gitops.git"
		u, err := url.Parse(normalizedURL)
		Expect(err).NotTo(HaveOccurred())
		_, err = getOwnerFromURL(*u, GitProviderGitLab)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("a subgroup in a subgroup is not currently supported"))
	})
})

type expectedRepoURL struct {
	s        string
	owner    string
	name     string
	provider GitProviderName
	protocol RepositoryURLProtocol
}

var _ = DescribeTable("NewRepoURL", func(input, gitProviderEnv string, expected expectedRepoURL) {
	if gitProviderEnv != "" {
		viper.Set("git-host-types", gitProviderEnv)
	}
	result, err := NewRepoURL(input)
	Expect(err).NotTo(HaveOccurred())

	Expect(result.String()).To(Equal(expected.s))
	u, err := url.Parse(expected.s)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.URL()).To(Equal(u))
	Expect(result.Owner()).To(Equal(expected.owner))
	Expect(result.Provider()).To(Equal(expected.provider))
	Expect(result.Protocol()).To(Equal(expected.protocol))
},
	Entry("github git clone style", "git@github.com:someuser/podinfo.git", "", expectedRepoURL{
		s:        "ssh://git@github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("github url style", "ssh://git@github.com/someuser/podinfo.git", "", expectedRepoURL{
		s:        "ssh://git@github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("github https", "https://github.com/someuser/podinfo.git", "", expectedRepoURL{
		s:        "ssh://git@github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("gitlab git clone style", "git@gitlab.com:someuser/podinfo.git", "", expectedRepoURL{
		s:        "ssh://git@gitlab.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitLab,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("gitlab https", "https://gitlab.com/someuser/podinfo.git", "", expectedRepoURL{
		s:        "ssh://git@gitlab.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitLab,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("trailing slash in url", "https://github.com/sympatheticmoose/podinfo-deploy/", "", expectedRepoURL{
		s:        "ssh://git@github.com/sympatheticmoose/podinfo-deploy.git",
		owner:    "sympatheticmoose",
		name:     "podinfo-deploy",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry(
		"custom domain",
		"git@gitlab.acme.org/sympatheticmoose/podinfo-deploy/",
		"gitlab.acme.org=gitlab",
		expectedRepoURL{
			s:        "ssh://git@gitlab.acme.org/sympatheticmoose/podinfo-deploy.git",
			owner:    "sympatheticmoose",
			name:     "podinfo-deploy",
			provider: "gitlab",
			protocol: RepositoryURLProtocolSSH,
		}),
)
