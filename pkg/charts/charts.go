package charts

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"sort"

	sourcev1beta1 "github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

// ProfileAnnotation is the annotation that Helm charts must have to indicate
// that they provide a Profile.
const ProfileAnnotation = "weave.works/profile"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . ChartScanner
type ChartScanner interface {
	ScanCharts(ctx context.Context, hr *sourcev1beta1.HelmRepository, pred ChartPredicate) ([]*pb.Profile, error)
}

type Scanner struct{}

// DefaultChartGetter provides default ways to get a chart index.yaml based on
// the URL scheme.
var DefaultChartGetters = getter.Providers{
	getter.Provider{
		Schemes: []string{"http", "https"},
		New:     getter.NewHTTPGetter,
	},
}

type ChartPredicate func(*repo.ChartVersion) bool

// Profiles is a predicate for scanning charts with the ProfileAnnotation.
var Profiles = func(v *repo.ChartVersion) bool {
	return hasAnnotation(v.Metadata, ProfileAnnotation)
}

// ScanCharts filters charts using the provided predicate.
//
// TODO: Add caching based on the Status Artifact Revision.

func (s *Scanner) ScanCharts(ctx context.Context, hr *sourcev1beta1.HelmRepository, pred ChartPredicate) ([]*pb.Profile, error) {
	chartRepo, err := fetchIndexFile(hr.Status.URL)
	if err != nil {
		return nil, fmt.Errorf("fetching profiles from HelmRepository %s/%s %q: %w",
			hr.GetName(), hr.GetNamespace(), hr.Spec.URL, err)
	}

	ps := make(map[string]*pb.Profile)
	for name, versions := range chartRepo.Entries {
		for _, v := range versions {
			if pred(v) {
				// if already added, update the versions array
				if p, ok := ps[name]; ok {
					p.AvailableVersions = append(p.AvailableVersions, v.Version)
				} else { // otherwise create a new profile and add to map
					p = &pb.Profile{
						Name:        name,
						Home:        v.Home,
						Sources:     v.Sources,
						Description: v.Description,
						Keywords:    v.Keywords,
						Icon:        v.Icon,
						KubeVersion: v.KubeVersion,
					}
					for _, m := range v.Maintainers {
						p.Maintainers = append(p.Maintainers, &pb.Maintainer{
							Name:  m.Name,
							Email: m.Email,
							Url:   m.URL,
						})
					}
					p.AvailableVersions = append(p.AvailableVersions, v.Version)
					ps[name] = p
				}
			}
		}
	}

	profiles := []*pb.Profile{}
	for _, p := range ps {
		sort.Strings(p.AvailableVersions)
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func fetchIndexFile(chartURL string) (*repo.IndexFile, error) {
	if hostname := os.Getenv("SOURCE_CONTROLLER_LOCALHOST"); hostname != "" {
		u, err := url.Parse(chartURL)
		if err != nil {
			return nil, err
		}
		u.Host = hostname
		chartURL = u.String()
	}

	u, err := url.Parse(chartURL)
	if err != nil {
		return nil, err
	}
	c, err := DefaultChartGetters.ByScheme(u.Scheme)
	if err != nil {
		return nil, fmt.Errorf("no provider for scheme: %s", u.Scheme)
	}

	res, err := c.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("get chart URL: %w", err)
	}

	b, err := ioutil.ReadAll(res)
	if err != nil {
		return nil, fmt.Errorf("read chart response: %w", err)
	}
	i := &repo.IndexFile{}
	if err := yaml.Unmarshal(b, i); err != nil {
		return nil, fmt.Errorf("unmarshaling chart response: %w", err)
	}
	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}

	i.SortEntries()

	return i, nil
}

func hasAnnotation(cm *chart.Metadata, name string) bool {
	for k := range cm.Annotations {
		if k == name {
			return true
		}
	}
	return false
}
