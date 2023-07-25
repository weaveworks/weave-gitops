package server

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	prefixWeight = 1000
	alphaWeight  = 1
	betaWeight   = 500
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

func versionRank(version string) (int, error) {
	if version == "__internal" {
		return 0, nil
	}

	re := regexp.MustCompile(`v(\d+)(alpha(\d+)|beta(\d+)|)$`)
	match := re.FindStringSubmatch(version)
	if match == nil {
		return -1, fmt.Errorf("invalid version string: %s", version)
	}

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

	return prefix*prefixWeight + suffixRank, nil
}

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

func DefaultPrimaryKinds() (*PrimaryKinds, error) {
	kinds := New()
	scheme, err := kube.CreateScheme()

	if err != nil {
		return nil, err
	}

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

// Add sets another kind name and gvk to resolve an object
// This errors if the kind is already set, as this likely indicates 2
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
