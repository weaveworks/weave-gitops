package install

import (
	"github.com/onsi/gomega"
	"testing"
)

func TestParseImageRepository(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tests := []struct {
		input    string
		repo     string
		img      string
		tag      string
		hasError bool
	}{
		{"localhost:5005/image:tag", "localhost:5005", "image", "tag", false},
		{"localhost:5005/org/image:tag", "localhost:5005/org", "image", "tag", false},
		{"localhost:5005/org/image", "localhost:5005/org", "image", "latest", false},
		{"org/sub-org/image:tag", "org/sub-org", "image", "tag", false},
		{"org/image:tag", "org", "image", "tag", false},
		{"org/image", "org", "image", "latest", false},
		{"image:tag", "", "image", "tag", false},
		{"image", "", "image", "", false},
		{"invalid:", "", "invalid", "", true},
		{":invalid", "", "", "invalid", true},
	}

	for _, test := range tests {
		repo, img, tag, err := parseImageRepository(test.input)
		g.Expect(repo).To(gomega.Equal(test.repo))
		g.Expect(img).To(gomega.Equal(test.img))
		g.Expect(tag).To(gomega.Equal(test.tag))
		if test.hasError {
			g.Expect(err).To(gomega.HaveOccurred())
		} else {
			g.Expect(err).NotTo(gomega.HaveOccurred())
		}
	}
}
