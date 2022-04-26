package profiles

import (
	"context"
	"fmt"
	"io"
	"strings"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/controller"
)

type ProfilesRetriever interface {
	Source() string
	RetrieveProfiles() (*pb.GetProfilesResponse, error)
}

type GetOptions struct {
	Name      string
	Version   string
	Cluster   string
	Namespace string
	Writer    io.Writer
	Port      string
}

func (s *ProfilesSvc) Get(ctx context.Context, r ProfilesRetriever, w io.Writer) error {
	profiles, err := r.RetrieveProfiles()
	if err != nil {
		return fmt.Errorf("unable to retrieve profiles from %q: %w", r.Source(), err)
	}

	printProfiles(profiles, w)

	return nil
}

// GetProfile returns a single available profile.
func (s *ProfilesSvc) GetProfile(ctx context.Context, r ProfilesRetriever, opts GetOptions) (*pb.Profile, string, error) {
	s.Logger.Actionf("getting available profiles from %s", r.Source())

	profilesList, err := r.RetrieveProfiles()
	if err != nil {
		return nil, "", fmt.Errorf("unable to retrieve profiles from %q: %w", r.Source(), err)
	}
	var version string

	for _, p := range profilesList.Profiles {
		if p.Name == opts.Name {
			if len(p.AvailableVersions) == 0 {
				return nil, "", fmt.Errorf("no version found for profile '%s' in %s/%s", p.Name, opts.Cluster, opts.Namespace)
			}

			switch {
			case opts.Version == "latest":
				versions, err := controller.ConvertStringListToSemanticVersionList(p.AvailableVersions)
				if err != nil {
					return nil, "", err
				}

				controller.SortVersions(versions)
				version = versions[0].String()
			default:
				if !foundVersion(p.AvailableVersions, opts.Version) {
					return nil, "", fmt.Errorf("version '%s' not found for profile '%s' in %s/%s", opts.Version, opts.Name, opts.Cluster, opts.Namespace)
				}

				version = opts.Version
			}

			if p.GetHelmRepository().GetName() == "" || p.GetHelmRepository().GetNamespace() == "" {
				return nil, "", fmt.Errorf("HelmRepository's name or namespace is empty")
			}

			return p, version, nil
		}
	}

	return nil, "", fmt.Errorf("no available profile '%s' found in %s/%s", opts.Name, opts.Cluster, opts.Namespace)
}

func foundVersion(availableVersions []string, version string) bool {
	for _, v := range availableVersions {
		if v == version {
			return true
		}
	}

	return false
}

func printProfiles(profiles *pb.GetProfilesResponse, w io.Writer) {
	fmt.Fprintf(w, "NAME\tDESCRIPTION\tAVAILABLE_VERSIONS\n")

	if profiles.Profiles != nil && len(profiles.Profiles) > 0 {
		for _, p := range profiles.Profiles {
			fmt.Fprintf(w, "%s\t%s\t%v", p.Name, p.Description, strings.Join(p.AvailableVersions, ","))
			fmt.Fprintln(w, "")
		}
	}
}
