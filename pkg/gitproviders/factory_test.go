package gitproviders

import (
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type expectedGitProvider struct {
	clientDomain     string
	clientProviderID string
	hostname         string
}

var _ = DescribeTable("buildGitProvider", func(input Config, expected expectedGitProvider) {
	c, h, err := buildGitProvider(input)
	Expect(err).ToNot(HaveOccurred())
	Expect(c.ProviderID()).To(Equal(gitprovider.ProviderID(expected.clientProviderID)), "ProviderID")
	Expect(c.SupportedDomain()).To(Equal(expected.clientDomain), "SupportedDomain")
	Expect(h).To(Equal(expected.hostname), "hostname")

},
	Entry("github.com", Config{Provider: "github", Hostname: "github.com", Token: "abc"}, expectedGitProvider{
		clientDomain:     "github.com",
		clientProviderID: "github",
		hostname:         "github.com",
	}),
	Entry("gitlab.com", Config{Provider: "gitlab", Hostname: "gitlab.com", Token: "abc"}, expectedGitProvider{
		// QUIRK..
		clientDomain:     "https://gitlab.com",
		clientProviderID: "gitlab",
		hostname:         "gitlab.com",
	}),
	Entry("github.acme.com", Config{Provider: "github", Hostname: "github.acme.com", Token: "abc"}, expectedGitProvider{
		clientDomain:     "https://github.acme.com",
		clientProviderID: "github",
		hostname:         "https://github.acme.com",
	}),
	Entry("gitlab.acme.com", Config{Provider: "gitlab", Hostname: "gitlab.acme.com", Token: "abc"}, expectedGitProvider{
		clientDomain:     "https://gitlab.acme.com",
		clientProviderID: "gitlab",
		hostname:         "https://gitlab.acme.com",
	}),
)
