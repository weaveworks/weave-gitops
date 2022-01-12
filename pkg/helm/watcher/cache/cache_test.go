package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
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

func TestCacheGetProfiles(t *testing.T) {
	t.Log("setup the cache and data for the cache")

	dir, err := os.MkdirTemp("", "cache-get-profiles")
	assert.NoError(t, err, "creating a temporary folder should have succeeded")

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Log("failed to cleanup the test folder")
		}
	}()

	helmCache, err := NewCache(dir)
	assert.NoError(t, err, "creating a new cache should have succeeded")

	t.Log("setting up data which contains multiple profiles including values")

	data := Data{
		Profiles: []*pb.Profile{profile1, profile2},
		Values: ValueMap{
			profile1.Name: values1,
			profile2.Name: values2,
		},
	}

	t.Log("call Update")
	assert.NoError(t, helmCache.Update(helmNamespace, helmName, data), "update call from cache should have worked")

	t.Log("call GetProfiles")

	profiles, err := helmCache.GetProfiles(helmNamespace, helmName)
	assert.NoError(t, err, "GetProfiles should not have run on an error")
	assert.Contains(t, profiles, profile1)
	assert.Contains(t, profiles, profile2)
}

func TestCacheGetProfilesNotFound(t *testing.T) {
	t.Log("setup the cache and data for the cache")

	dir, err := os.MkdirTemp("", "cache-get-profiles")
	assert.NoError(t, err, "creating a temporary folder should have succeeded")

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Log("failed to cleanup the test folder")
		}
	}()

	helmCache, err := NewCache(dir)
	assert.NoError(t, err, "creating a new cache should have succeeded")

	t.Log("setting up data which contains a single profile without values")

	data := Data{
		Profiles: []*pb.Profile{profile1},
	}

	t.Log("call Update")
	assert.NoError(t, helmCache.Update(helmNamespace, helmName, data), "update call from cache should have worked")

	t.Log("call GetProfiles")

	_, err = helmCache.GetProfiles("not-found", "none")
	assert.EqualError(t, err, fmt.Sprintf("failed to read profiles data for helm repo: open %s: no such file or directory", filepath.Join(dir, "not-found", "none", profileFilename)))
}

func TestCacheGetProfilesInvalidDataInFile(t *testing.T) {
	t.Log("setup the cache and data for the cache")

	dir, err := os.MkdirTemp("", "cache-get-profiles")
	assert.NoError(t, err, "creating a temporary folder should have succeeded")

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Log("failed to cleanup the test folder")
		}
	}()

	helmCache, err := NewCache(dir)
	assert.NoError(t, err, "creating a new cache should have succeeded")

	t.Log("setting up data which contains a single profile without values")

	data := Data{
		Profiles: []*pb.Profile{profile1},
	}

	t.Log("call Update")
	assert.NoError(t, helmCache.Update(helmNamespace, helmName, data), "update call from cache should have worked")

	t.Log("corrupt the profiles file so it's no longer valid yaml")
	assert.NoError(t, os.WriteFile(filepath.Join(dir, helmNamespace, helmName, profileFilename), []byte("empty"), 0700))

	t.Log("call GetProfiles for the corrupted profiles.yaml file")

	_, err = helmCache.GetProfiles(helmNamespace, helmName)
	assert.EqualError(t, err, "failed to unmarshal profiles data: error unmarshaling JSON: json: cannot unmarshal string into Go value of type []*profiles.Profile")
}

func TestCacheGetProfileValues(t *testing.T) {
	t.Log("setup the cache and data for the cache")

	dir, err := os.MkdirTemp("", "cache-get-profiles")
	assert.NoError(t, err, "creating a temporary folder should have succeeded")

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Log("failed to cleanup the test folder")
		}
	}()

	helmCache, err := NewCache(dir)
	assert.NoError(t, err, "creating a new cache should have succeeded")

	t.Log("setting up data which contains multiple profiles including values")

	data := Data{
		Profiles: []*pb.Profile{profile1, profile2},
		Values: ValueMap{
			profile1.Name: values1,
			profile2.Name: values2,
		},
	}

	t.Log("call Update")
	assert.NoError(t, helmCache.Update(helmNamespace, helmName, data), "update call from cache should have worked")

	t.Log("call GetProfileValues with profile1")

	value, err := helmCache.GetProfileValues(helmNamespace, helmName, profile1.Name, "0.0.2")
	assert.NoError(t, err)
	assert.Equal(t, []byte("values-2"), value)

	t.Log("call GetProfileValues with profile2")

	value, err = helmCache.GetProfileValues(helmNamespace, helmName, profile2.Name, "0.0.5")
	assert.NoError(t, err)
	assert.Equal(t, []byte("values-5"), value)
}

func TestGetProfileValuesNonexistent(t *testing.T) {
	t.Log("setup the cache and data for the cache")

	dir, err := os.MkdirTemp("", "cache-get-profiles")
	assert.NoError(t, err, "creating a temporary folder should have succeeded")

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Log("failed to cleanup the test folder")
		}
	}()

	helmCache, err := NewCache(dir)
	assert.NoError(t, err, "creating a new cache should have succeeded")

	t.Log("setting up data which contains a single profile and values")

	data := Data{
		Profiles: []*pb.Profile{profile1},
		Values: ValueMap{
			profile1.Name: values1,
		},
	}

	t.Log("call Update")
	assert.NoError(t, helmCache.Update(helmNamespace, helmName, data), "update call from cache should have worked")

	t.Log("call GetProfileValues with nonexistent values version")

	_, err = helmCache.GetProfileValues(helmNamespace, helmName, profile1.Name, "999")
	assert.EqualError(t, err, fmt.Sprintf("failed to read values file: open %s/test-namespace/test-name/test-profiles-1/999/values.yaml: no such file or directory", dir))
}

func TestGetProfilesFailedLock(t *testing.T) {
	helmCache := &HelmCache{cacheDir: "nope"}
	_, err := helmCache.GetProfiles("", "")
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}

func TestGetProfileValuesFailedLock(t *testing.T) {
	helmCache := &HelmCache{cacheDir: "nope"}
	_, err := helmCache.GetProfileValues("", "", "", "")
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}

func TestUpdateFailedLock(t *testing.T) {
	helmCache := &HelmCache{cacheDir: "nope"}
	err := helmCache.Update("", "", Data{})
	assert.EqualError(t, err, "unable to read lock file cache.lock: open nope/cache.lock: no such file or directory")
}
