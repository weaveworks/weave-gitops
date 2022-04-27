package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/gofrs/flock"

	pb "github.com/weaveworks/weave-gitops/gitops/pkg/api/profiles"
)

const (
	// we globally lock the whole base cache folder to avoid the
	// possibility of checking if a sub-folder exists, just so it is created the moment
	// we try to create it after we checked that it exists.
	lockFilename    = "cache.lock"
	profileFilename = "profiles.yaml"
	valuesFilename  = "values.yaml"
	lockTimeout     = 1 * time.Minute
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
	Put(ctx context.Context, helmRepoNamespace, helmRepoName string, value Data) error
	Delete(ctx context.Context, helmRepoNamespace, helmRepoName string) error
	// ListProfiles specifically retrieve profiles data only to avoid traversing the values structure for no reason.
	ListProfiles(ctx context.Context, helmRepoNamespace, helmRepoName string) ([]*pb.Profile, error)
	// GetProfileValues will try and find a specific values file for the given profileName and profileVersion. Returns an
	// error if said version is not found.
	GetProfileValues(ctx context.Context, helmRepoNamespace, helmRepoName, profileName, profileVersion string) ([]byte, error)
	// ListAvailableVersionsForProfile returns all stored available versions for a profile.
	ListAvailableVersionsForProfile(ctx context.Context, helmRepoNamespace, helmRepoName, profileName string) ([]string, error)
}

// Data is explicit data for a specific profile including values.
// Saved as `profiles.yaml` and `profileName/version/values.yaml`.
type Data struct {
	Profiles []*pb.Profile `yaml:"profiles"`
	Values   ValueMap
}

// ProfileCache is used to cache profiles data from scanner helm repositories.
type ProfileCache struct {
	cacheLocation string
}

var _ Cache = &ProfileCache{}

// NewCache initialises the cache and returns it.
func NewCache(cacheLocation string) (*ProfileCache, error) {
	if err := os.MkdirAll(cacheLocation, 0700); err != nil {
		return nil, fmt.Errorf("failed to create helm cache dir: %w", err)
	}

	return &ProfileCache{
		cacheLocation: cacheLocation,
	}, nil
}

// Put adds a new entry or updates an existing entry in the cache for the helmRepository.
func (c *ProfileCache) Put(ctx context.Context, helmRepoNamespace, helmRepoName string, value Data) error {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("starting put operation")

	putOperation := func() error {
		// namespace and name are already sanitized and should not be able to pass in
		// things like `../` and `../usr/`, etc.
		cacheLocation := filepath.Join(c.cacheLocation, helmRepoNamespace, helmRepoName)
		if err := os.MkdirAll(cacheLocation, 0700); err != nil {
			return fmt.Errorf("failed to create cache location: %w", err)
		}

		// write out the `profiles` data as yaml
		profileData, err := yaml.Marshal(value.Profiles)
		if err != nil {
			return fmt.Errorf("failed to marshal profile data: %w", err)
		}

		if err := os.WriteFile(filepath.Join(cacheLocation, profileFilename), profileData, 0700); err != nil {
			return fmt.Errorf("failed to write profile data: %w", err)
		}

		// create the `values` folder structure and write out the `values` data
		for profName, versions := range value.Values {
			for version, values := range versions {
				versionFolder := filepath.Join(cacheLocation, profName, version)

				if err := os.MkdirAll(versionFolder, 0700); err != nil {
					return fmt.Errorf("failed to create version folder %s for profile %s: %w", version, profName, err)
				}

				if err := os.WriteFile(filepath.Join(versionFolder, valuesFilename), values, 0700); err != nil {
					return fmt.Errorf("failed to write out values for version %s: %w", version, err)
				}
			}
		}

		logger.Info("finished put operation")

		return nil
	}

	return c.tryWithLock(ctx, putOperation)
}

// Delete clears the cache folder for a specific HelmRepository. It will only clear the innermost
// folder so others in the same namespace may retain their values.
func (c *ProfileCache) Delete(ctx context.Context, helmRepoNamespace, helmRepoName string) error {
	deleteOperation := func() error {
		location := filepath.Join(c.cacheLocation, helmRepoNamespace, helmRepoName)
		if err := os.RemoveAll(location); err != nil {
			return fmt.Errorf("failed to clean up cache for location %s with error: %w", location, err)
		}

		return nil
	}

	return c.tryWithLock(ctx, deleteOperation)
}

// ListProfiles gathers all profiles for a helmRepo if found. Returns an error otherwise.
func (c *ProfileCache) ListProfiles(ctx context.Context, helmRepoNamespace, helmRepoName string) ([]*pb.Profile, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("retrieving cached profile data")

	var result []*pb.Profile

	listOperation := func() error {
		if err := c.getProfilesFromFile(helmRepoNamespace, helmRepoName, &result); err != nil {
			return fmt.Errorf("failed to read profiles data for helm repo (%s/%s): %w", helmRepoNamespace, helmRepoName, err)
		}

		return nil
	}

	if err := c.tryWithLock(ctx, listOperation); err != nil {
		return nil, err
	}

	return result, nil
}

// ListAvailableVersionsForProfile returns all stored available versions for a profile.
func (c *ProfileCache) ListAvailableVersionsForProfile(ctx context.Context, helmRepoNamespace, helmRepoName, profileName string) ([]string, error) {
	// Because the folders of versions are only stored when there are values for a version
	// we, instead, look in the profiles.yaml file for available versions.
	var result []string

	getAllAvailableVersionsOp := func() error {
		var profiles []*pb.Profile
		if err := c.getProfilesFromFile(helmRepoNamespace, helmRepoName, &profiles); err != nil {
			if os.IsNotExist(err) {
				return nil
			}

			return fmt.Errorf("failed to read profiles data for helm repo: %w", err)
		}

		for _, p := range profiles {
			if p.Name == profileName {
				result = append(result, p.AvailableVersions...)
				return nil
			}
		}

		return fmt.Errorf("profile with name %s not found in cached profiles", profileName)
	}

	if err := c.tryWithLock(ctx, getAllAvailableVersionsOp); err != nil {
		return nil, err
	}

	return result, nil
}

// GetProfileValues returns the content of the cached values file if it exists. Errors otherwise.
func (c *ProfileCache) GetProfileValues(ctx context.Context, helmRepoNamespace, helmRepoName, profileName, profileVersion string) ([]byte, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("retrieving cached profile values data")

	var result []byte

	getValuesOperation := func() error {
		values, err := os.ReadFile(filepath.Join(c.cacheLocation, helmRepoNamespace, helmRepoName, profileName, profileVersion, valuesFilename))
		if err != nil {
			return fmt.Errorf("failed to read values file: %w", err)
		}

		result = values

		return nil
	}

	if err := c.tryWithLock(ctx, getValuesOperation); err != nil {
		return nil, err
	}

	return result, nil
}

// getProfilesFromFile returns profiles loaded from a file.
func (c *ProfileCache) getProfilesFromFile(helmRepoNamespace, helmRepoName string, profiles *[]*pb.Profile) error {
	content, err := os.ReadFile(filepath.Join(c.cacheLocation, helmRepoNamespace, helmRepoName, profileFilename))
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(content, profiles); err != nil {
		return err
	}

	return nil
}

// tryWithLock tries to run the given operation by acquiring a lock first.
func (c *ProfileCache) tryWithLock(ctx context.Context, operationFunc func() error) error {
	lock := flock.New(filepath.Join(c.cacheLocation, lockFilename))

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Printf("Unable to unlock file %s: %v\n", lockFilename, err)
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, lockTimeout)
	defer cancel()

	ok, err := lock.TryRLockContext(ctx, 250*time.Millisecond)
	if !ok {
		return fmt.Errorf("unable to read lock file %s: %w", lockFilename, err)
	}

	return operationFunc()
}
