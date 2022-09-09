package helm

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Service represents the elements that we need to use the client Proxy to fetch
// a URL.
type Service struct {
	Scheme    string
	Namespace string
	Name      string
	Path      string
	Port      string
}

// ParseArtifactURL takes HelmRepository Artifact URL for a remote cluster and
// returns the components of the URL.
func ParseArtifactURL(artifactURL string) (*Service, error) {
	u, err := url.Parse(artifactURL)
	if err != nil {
		return nil, err
	}

	// Split hostname to get namespace and name.
	host := strings.Split(u.Hostname(), ".")

	if len(host) != 5 || host[2] != "svc" || u.Path == "/" {
		return nil, fmt.Errorf("invalid artifact URL %s", artifactURL)
	}

	port := u.Port()
	if port == "" {
		port = "80"
	}

	return &Service{
		Scheme:    u.Scheme,
		Namespace: host[1],
		Name:      host[0],
		Path:      u.Path,
		Port:      port,
	}, nil
}

func TestParseService(t *testing.T) {
	testCases := []struct {
		name        string
		artifactURL string
		want        *Service
		err         error
	}{
		{
			"parses correctly",
			"http://source-controller.flux-system.svc.cluster.local/demo-index.yaml",
			&Service{
				Scheme:    "http",
				Namespace: "flux-system",
				Name:      "source-controller",
				Path:      "/demo-index.yaml",
				Port:      "80",
			},
			nil,
		},
		{
			"wrong url",
			"http://github.com/example.repo",
			nil,
			errors.New("invalid artifact URL http://github.com/example.repo"),
		},
		{
			"empty path",
			"http://source-controller.flux-system.svc.cluster.local/",
			nil,
			errors.New("invalid artifact URL http://source-controller.flux-system.svc.cluster.local/"),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseArtifactURL(tt.artifactURL)
			if tt.err != nil {
				if diff := cmp.Diff(tt.err.Error(), err.Error()); diff != "" {
					t.Fatalf("got wrong error:\n%s", diff)
				}
			}

			if diff := cmp.Diff(tt.want, parsed); diff != "" {
				t.Fatalf("failed to parse URL:\n%s", diff)
			}
		})
	}
}
