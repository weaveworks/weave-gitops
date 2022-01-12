package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gofrs/flock"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
)

const (
	// we globally lock the whole base cache folder to avoid the
	// possibility of checking if a sub-folder exists, just so it is created the moment
	// we try to create it after we checked that it exists.
	lockFilename    = "cache.lock"
	profileFilename = "profiles.yaml"
	valuesFilename  = "values.yaml"
)

// type alias for easier reading of the cache layout for values.
type profileName = string
type profileVersion = string

// ValueMap contains easy access for a profile name and version based values file.
type ValueMap map[profileName]map[profileVersion][]byte

// Cache defines an interface to work with the profile data cacher.
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Cache
type Cache interface {
	Update(helmRepoNamespace, helmRepoName string, value Data) error
	// GetProfiles specifically retrieve profiles data only to avoid traversing the values structure for no reason.
	GetProfiles(helmRepoNamespace, helmRepoName string) ([]*pb.Profile, error)
	// GetProfileValues will try and find a specific values file for the given profileName and profileVersion. Returns an
	// error if said version is not found.
	GetProfileValues(helmRepoNamespace, helmRepoName, profileName, profileVersion string) ([]byte, error)
}

// Data is explicit data for a specific profile including values.
// Saved as `profiles.yaml` and `profileName/version/values.yaml`.
type Data struct {
	Profiles []*pb.Profile `yaml:"profiles"`
	Values   ValueMap
}

// HelmCache is used to cache profiles data from scanner helm repositories.
type HelmCache struct {
	cacheDir string
}

var _ Cache = &HelmCache{}

// NewCache initialises the cache and returns it.
func NewCache(cacheDir string) (*HelmCache, error) {
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create helm cache dir: %w", err)
		}
	}

	return &HelmCache{
		cacheDir: cacheDir,
	}, nil
}

// Update adds a new entry to the cache for the current helmRepository.
func (c *HelmCache) Update(helmRepoNamespace, helmRepoName string, value Data) error {
	// acquire lock
	lock := flock.New(filepath.Join(c.cacheDir, lockFilename))
	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Printf("Unable to unlock file %s: %v\n", lockFilename, err)
		}
	}()

	// release on timeout
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	ok, err := lock.TryRLockContext(ctx, 250*time.Millisecond) // try to lock every 1/4 second
	if !ok {
		return fmt.Errorf("unable to read lock file %s: %w", lockFilename, err)
	}

	// namespace and name are already sanitized and should not be able to pass in
	// things like `../` and `../usr/`, etc.
	cacheFolder := filepath.Join(c.cacheDir, helmRepoNamespace, helmRepoName)
	if err := os.MkdirAll(cacheFolder, 0700); err != nil {
		return fmt.Errorf("failed to create cache folder: %w", err)
	}

	// write out the `profiles` data as yaml
	profileData, err := yaml.Marshal(value.Profiles)
	if err != nil {
		return fmt.Errorf("failed to marshal profile data: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cacheFolder, profileFilename), profileData, 0700); err != nil {
		return fmt.Errorf("failed to write profile data: %w", err)
	}

	// create the `values` folder structure and write out the `values` data
	for profName, versions := range value.Values {
		for version, values := range versions {
			versionFolder := filepath.Join(cacheFolder, profName, version)

			if err := os.MkdirAll(versionFolder, 0700); err != nil {
				return fmt.Errorf("failed to create version folder %s for profile %s: %w", version, profName, err)
			}

			if err := os.WriteFile(filepath.Join(versionFolder, valuesFilename), values, 0700); err != nil {
				return fmt.Errorf("failed to write out values for version %s: %w", version, err)
			}
		}
	}

	return nil
}

// GetProfiles gathers all profiles for a helmRepo if found. Returns an error otherwise.
func (c *HelmCache) GetProfiles(helmRepoNamespace, helmRepoName string) ([]*pb.Profile, error) {
	lock := flock.New(filepath.Join(c.cacheDir, lockFilename))

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Printf("Unable to unlock file %s: %v\n", lockFilename, err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	ok, err := lock.TryRLockContext(ctx, 250*time.Millisecond)
	if !ok {
		return nil, fmt.Errorf("unable to read lock file %s: %w", lockFilename, err)
	}

	var result []*pb.Profile

	content, err := os.ReadFile(filepath.Join(c.cacheDir, helmRepoNamespace, helmRepoName, profileFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles data for helm repo: %w", err)
	}

	if err := yaml.Unmarshal(content, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profiles data: %w", err)
	}

	return result, nil
}

// GetProfileValues returns the content of the cached values file if it exists. Errors otherwise.
func (c *HelmCache) GetProfileValues(helmRepoNamespace, helmRepoName, profileName, profileVersion string) ([]byte, error) {
	lock := flock.New(filepath.Join(c.cacheDir, lockFilename))

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Printf("Unable to unlock file %s: %v\n", lockFilename, err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	ok, err := lock.TryRLockContext(ctx, 250*time.Millisecond)
	if !ok {
		return nil, fmt.Errorf("unable to read lock file %s: %w", lockFilename, err)
	}

	values, err := os.ReadFile(filepath.Join(c.cacheDir, helmRepoNamespace, helmRepoName, profileName, profileVersion, valuesFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to read values file: %w", err)
	}

	return values, nil
}
