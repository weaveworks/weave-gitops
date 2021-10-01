package gitproviders

import (
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("DetectGitProviderFromUrl", func(input string, expected GitProviderName) {
	result, err := DetectGitProviderFromUrl(input)
	Expect(err).NotTo(HaveOccurred())
	Expect(result).To(Equal(expected))
},
	Entry("ssh+github", "ssh://git@github.com/weaveworks/weave-gitops.git", GitProviderGitHub),
	Entry("ssh+gitlab", "ssh://git@gitlab.com/weaveworks/weave-gitops.git", GitProviderGitLab),
)

type expectedRepoURL struct {
	s        string
	owner    string
	name     string
	provider GitProviderName
	protocol RepositoryURLProtocol
}

var _ = DescribeTable("NormalizedRepoURL", func(input string, expected expectedRepoURL) {
	result, err := NewNormalizedRepoURL(input)
	Expect(err).NotTo(HaveOccurred())

	Expect(result.String()).To(Equal(expected.s))
	u, err := url.Parse(expected.s)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.URL()).To(Equal(u))
	Expect(result.Owner()).To(Equal(expected.owner))
	Expect(result.Provider()).To(Equal(expected.provider))
	Expect(result.Protocol()).To(Equal(expected.protocol))
},
	Entry("github git clone style", "git@github.com:someuser/podinfo.git", expectedRepoURL{
		s:        "ssh://git@github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("github url style", "ssh://git@github.com/someuser/podinfo.git", expectedRepoURL{
		s:        "ssh://git@github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("github https", "https://github.com/someuser/podinfo.git", expectedRepoURL{
		s:        "https://github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolHTTPS,
	}),
	Entry("gitlab git clone style", "git@gitlab.com:someuser/podinfo.git", expectedRepoURL{
		s:        "ssh://git@gitlab.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitLab,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("gitlab https", "https://gitlab.com/someuser/podinfo.git", expectedRepoURL{
		s:        "https://gitlab.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitLab,
		protocol: RepositoryURLProtocolHTTPS,
	}),
)

var _ = Describe("Test GetRepoVisiblity", func() {
	// url := "ssh://git@github.com/foo/bar"
	// It("tests that a nil info generates the appropriate error", func() {
	// 	result, underlyingError := getVisibilityFromRepoInfo(url, nil)
	// 	Expect(result).To(BeNil())
	// 	Expect(underlyingError.Error()).To(Equal(fmt.Sprintf("unable to obtain repository visibility for: %s", url)))
	// })

	// It("tests that a nil visibility reference generates the appropriate error", func() {
	// 	result, underlyingError := getVisibilityFromRepoInfo(url, &gitprovider.RepositoryInfo{Visibility: nil})
	// 	Expect(result).To(BeNil())
	// 	Expect(underlyingError.Error()).To(Equal(fmt.Sprintf("unable to obtain repository visibility for: %s", url)))
	// })

	// It("tests that a non-nil visibility reference is successful", func() {
	// 	public := gitprovider.RepositoryVisibilityPublic
	// 	result, underlyingError := getVisibilityFromRepoInfo(url, &gitprovider.RepositoryInfo{Visibility: &public})
	// 	Expect(underlyingError).To(BeNil())
	// 	Expect(result).To(Equal(&public))
	// })
})
