package server

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// prefixWeight is a weight assigned to the major version number (the digit after "v" in a version string).
	// This weight is used in calculating the ranking score of a version.
	// Higher values give more priority to the major version number.
	prefixWeight = 1000

	// alphaWeight is a weight assigned to alpha versions in the version string.
	// It's multiplied with the digit following the "alpha" suffix.
	// Alpha versions are early pre-release software and typically contain bugs, hence they have a lower weight.
	alphaWeight = 1

	// betaWeight is a weight assigned to beta versions in the version string.
	// It's multiplied with the digit following the "beta" suffix.
	// Beta versions are more stable than alpha versions but might still contain bugs.
	// They are given a higher weight than alpha versions.
	betaWeight = 500

	// stableWeight is a weight assigned to stable versions (those without an "alpha" or "beta" suffix).
	// Stable versions are the final, production-ready releases of software, hence they have the highest weight.
	stableWeight = 1000
)

type PrimaryKinds struct {
	kinds map[string]schema.GroupVersionKind
}

func New() *PrimaryKinds {
	kinds := PrimaryKinds{}
	kinds.kinds = make(map[string]schema.GroupVersionKind)

	return &kinds
}

// versionRank calculates a ranking score for a given version string. It takes into account
// the prefix number (for example, in v1alpha2, 1 is the prefix) and the suffix
// (for example, alpha or beta and their accompanying number, if present).
// The function returns an integer representing the ranking score and any error that occurred.
func versionRank(version string) (int, error) {
	// special handling for the internal version
	if version == "__internal" {
		return 0, nil
	}

	// regex pattern to extract version information
	re := regexp.MustCompile(`v(\d+)(alpha(\d+)|beta(\d+)|)$`)
	match := re.FindStringSubmatch(version)
	if match == nil {
		return -1, fmt.Errorf("invalid version string: %s", version)
	}

	// get and process the version prefix and suffix
	prefix, _ := strconv.Atoi(match[1])
	suffixRank := 0
	if match[3] != "" {
		suffixRank, _ = strconv.Atoi(match[3])
		suffixRank *= alphaWeight
	} else if match[4] != "" {
		suffixRank, _ = strconv.Atoi(match[4])
		suffixRank *= betaWeight
	} else {
		suffixRank = stableWeight
	}

	// calculate and return the rank
	return prefix*prefixWeight + suffixRank, nil
}

// compareGVK compares two GroupVersionKind (GVK) objects and determines
// which has a higher version rank. The function returns an integer (-1, 0, or 1)
// representing the comparison result, and any error that occurred.
func compareGVK(gvk1, gvk2 schema.GroupVersionKind) (int, error) {
	rank1, err1 := versionRank(gvk1.Version)
	if err1 != nil {
		return 0, fmt.Errorf("failed to parse gvk group: %s, version %s, kind:%s : %w", gvk1.Group, gvk1.Version, gvk1.Kind, err1)
	}

	rank2, err2 := versionRank(gvk2.Version)
	if err2 != nil {
		return 0, err2
	}

	if rank1 > rank2 {
		return 1, nil
	} else if rank1 < rank2 {
		return -1, nil
	} else {
		return 0, nil
	}
}

type KnownTypes interface {
	AllKnownTypes() map[schema.GroupVersionKind]reflect.Type
}

// DefaultPrimaryKinds generates a new PrimaryKinds object which contains
// the highest version of each kind of known types from a Kubernetes scheme.
// It returns the PrimaryKinds object and any error that occurred.
func DefaultPrimaryKinds() (*PrimaryKinds, error) {
	scheme, err := kube.CreateScheme()
	if err != nil {
		return nil, err
	}

	return getPrimaryKinds(scheme)
}

// getPrimaryKinds generates a new PrimaryKinds object which contains
// the highest version of each kind of known types from a Kubernetes scheme.
// It is used internally by DefaultPrimaryKinds
// to return the PrimaryKinds object and any error that occurred.
func getPrimaryKinds(scheme KnownTypes) (*PrimaryKinds, error) {
	kinds := New()

	for gvk := range scheme.AllKnownTypes() {
		existedGvk, exist := kinds.kinds[gvk.Kind]
		if exist {
			compareResult, err := compareGVK(gvk, existedGvk)
			if err != nil {
				return nil, err
			}

			// gvk is larger than existedGvk, replace it
			if compareResult > 0 {
				kinds.kinds[gvk.Kind] = gvk
			}
		} else {
			// gvk is not existed, add it
			kinds.kinds[gvk.Kind] = gvk
		}
	}

	return kinds, nil
}

// Add sets another kind name and gvk to resolve an object.
// This function returns an error if the kind is already set, as this likely indicates 2
// different uses for the same kind string.
func (pk *PrimaryKinds) Add(kind string, gvk schema.GroupVersionKind) error {
	_, ok := pk.kinds[kind]
	if ok {
		return fmt.Errorf("couldn't add kind %v - already added", kind)
	}

	pk.kinds[kind] = gvk

	return nil
}

// Lookup ensures that a kind name is known, white-listed, and returns
// the full GVK for that kind
func (pk *PrimaryKinds) Lookup(kind string) (*schema.GroupVersionKind, error) {
	gvk, ok := pk.kinds[kind]
	if !ok {
		return nil, fmt.Errorf("looking up objects of kind %v not supported", kind)
	}

	return &gvk, nil
}
