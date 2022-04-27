package cache

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/weaveworks/weave-gitops/gitops/pkg/api/profiles"
)

var (
	profile1 = &pb.Profile{
		Name:              "test-profiles-1",
		Home:              "home",
		Description:       "description",
		Icon:              "icon",
		KubeVersion:       "1.21",
		AvailableVersions: []string{"0.0.1", "0.0.2", "0.0.3"},
	}
	profile2 = &pb.Profile{
		Name:              "test-profiles-2",
		Home:              "home",
		Description:       "description",
		Icon:              "icon",
		KubeVersion:       "1.21",
		AvailableVersions: []string{"0.0.4", "0.0.5", "0.0.6"},
	}
	values1 = map[profileVersion][]byte{
		"0.0.2": []byte("values-2"),
		"0.0.3": []byte("values-3"),
	}
	values2 = map[profileVersion][]byte{
		"0.0.5": []byte("values-5"),
	}
	helmNamespace = "test-namespace"
	helmName      = "test-name"
)

func TestCacheListProfiles(t *testing.T) {
	profileCache, _ := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1, profile2},
		Values: ValueMap{
			profile1.Name: values1,
			profile2.Name: values2,
		},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	profiles, err := profileCache.ListProfiles(context.Background(), helmNamespace, helmName)
	assert.NoError(t, err, "ListProfiles should not have run on an error")
	assert.Contains(t, profiles, profile1)
	assert.Contains(t, profiles, profile2)
}

func TestCacheListProfilesNotFound(t *testing.T) {
	profileCache, dir := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	_, err := profileCache.ListProfiles(context.Background(), "not-found", "none")
	assert.EqualError(t, err,
		fmt.Sprintf("failed to read profiles data for helm repo (%s/%s): open %s: no such file or directory",
			"not-found",
			"none",
			filepath.Join(dir, "not-found", "none", profileFilename)))
}

func TestCacheListProfilesInvalidDataInFile(t *testing.T) {
	profileCache, dir := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	assert.NoError(t, os.WriteFile(filepath.Join(dir, helmNamespace, helmName, profileFilename), []byte("empty"), 0700))
	_, err := profileCache.ListProfiles(context.Background(), helmNamespace, helmName)
	assert.EqualError(t, err,
		fmt.Sprintf("failed to read profiles data for helm repo (%s/%s): "+
			"error unmarshaling JSON: json: cannot unmarshal string into Go value of type []*profiles.Profile", helmNamespace, helmName))
}

func TestCacheGetProfileValues(t *testing.T) {
	profileCache, _ := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1, profile2},
		Values: ValueMap{
			profile1.Name: values1,
			profile2.Name: values2,
		},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	value, err := profileCache.GetProfileValues(context.Background(), helmNamespace, helmName, profile1.Name, "0.0.2")
	assert.NoError(t, err)
	assert.Equal(t, []byte("values-2"), value)
	value, err = profileCache.GetProfileValues(context.Background(), helmNamespace, helmName, profile2.Name, "0.0.5")
	assert.NoError(t, err)
	assert.Equal(t, []byte("values-5"), value)
}

func TestGetProfileValuesNonexistent(t *testing.T) {
	profileCache, dir := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1},
		Values: ValueMap{
			profile1.Name: values1,
		},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	_, err := profileCache.GetProfileValues(context.Background(), helmNamespace, helmName, profile1.Name, "999")
	assert.EqualError(t, err, fmt.Sprintf("failed to read values file: open %s/test-namespace/test-name/test-profiles-1/999/values.yaml: no such file or directory", dir))
}

// Note that error case is missing. It's actually difficult to make RemoveAll fail. It
// could fail if we mess up the permission on a file, but that would leave us with a file we can't
// clear up and neither modify.
func TestDeleteExistingData(t *testing.T) {
	profileCache, dir := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1},
		Values: ValueMap{
			profile1.Name: values1,
		},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	assert.NoError(t, profileCache.Delete(context.Background(), helmNamespace, helmName), "delete operation should have worked")
	_, err := os.Stat(filepath.Join(dir, helmNamespace, helmName))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestListAvailableVersionsForProfile(t *testing.T) {
	profileCache, _ := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1},
		Values: ValueMap{
			profile1.Name: values1,
		},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	versions, err := profileCache.ListAvailableVersionsForProfile(context.Background(), helmNamespace, helmName, profile1.Name)
	assert.NoError(t, err)
	assert.Equal(t, profile1.AvailableVersions, versions)
}

func TestListAvailableVersionsForProfileNoCachedData(t *testing.T) {
	profileCache, _ := setupCache(t)
	versions, err := profileCache.ListAvailableVersionsForProfile(context.Background(), helmNamespace, helmName, profile1.Name)
	assert.NoError(t, err)
	assert.Nil(t, versions)
}

func TestListAvailableVersionsForProfileNameNotFound(t *testing.T) {
	profileCache, _ := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1},
		Values: ValueMap{
			profile1.Name: values1,
		},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	_, err := profileCache.ListAvailableVersionsForProfile(context.Background(), helmNamespace, helmName, "notfound")
	assert.EqualError(t, err, "profile with name notfound not found in cached profiles")
}

func TestListAvailableVersionsForProfileInvalidYamlData(t *testing.T) {
	profileCache, dir := setupCache(t)
	data := Data{
		Profiles: []*pb.Profile{profile1},
		Values: ValueMap{
			profile1.Name: values1,
		},
	}
	assert.NoError(t, profileCache.Put(context.Background(), helmNamespace, helmName, data), "put call from cache should have worked")
	assert.NoError(t, os.WriteFile(filepath.Join(dir, helmNamespace, helmName, profileFilename), []byte("empty"), 0700))
	_, err := profileCache.ListAvailableVersionsForProfile(context.Background(), helmNamespace, helmName, profile1.Name)
	assert.EqualError(t, err, "failed to read profiles data for helm repo: error unmarshaling JSON: json: cannot unmarshal string into Go value of type []*profiles.Profile")
}

func TestListProfilesFailedLock(t *testing.T) {
	profileCache := &ProfileCache{cacheLocation: "nope"}
	_, err := profileCache.ListProfiles(context.Background(), "", "")
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}

func TestGetProfileValuesFailedLock(t *testing.T) {
	profileCache := &ProfileCache{cacheLocation: "nope"}
	_, err := profileCache.GetProfileValues(context.Background(), "", "", "", "")
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}

func TestUpdateFailedLock(t *testing.T) {
	profileCache := &ProfileCache{cacheLocation: "nope"}
	err := profileCache.Put(context.Background(), "", "", Data{})
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}

func TestDeleteFailedLock(t *testing.T) {
	profileCache := &ProfileCache{cacheLocation: "nope"}
	err := profileCache.Delete(context.Background(), "", "")
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}

func TestTestListAvailableVersionsForProfileFailedLock(t *testing.T) {
	profileCache := &ProfileCache{cacheLocation: "nope"}
	_, err := profileCache.ListAvailableVersionsForProfile(context.Background(), "", "", "")
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}

func setupCache(t *testing.T) (Cache, string) {
	dir, err := os.MkdirTemp("", "cache-temp-dir")
	assert.NoError(t, err, "creating a temporary folder should have succeeded")
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatalf("failed to cleanup the test folder: %s", err)
		}
	})

	profileCache, err := NewCache(dir)
	assert.NoError(t, err, "creating a new cache should have succeeded")

	return profileCache, dir
}
