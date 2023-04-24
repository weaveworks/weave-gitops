package install

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Heuristic mapping from the Source controller version to the Flux version.
// We support back to v0.32 only and the guess will be only the main version number.
// For example, Flux v0.38.x versions contain the Source controller v0.33.0,
// but we will report the Flux version as v0.38.0 only.
//
// Source: https://github.com/fluxcd/flux2/releases
//
// How to update this map:
// 1. Find the Flux version you want to support in the URL above.
// 2. Find the Source controller version that is used in that Flux version
// 3. Add the new mapping to the map below
// 4. Don't forget to *remove the oldest mapping* from the map below
var sourceVerToFluxVer = map[string]string{
	"v1.0.0-rc.1": "v2.0.0-rc.1",
	"v0.36.1":     "v0.41.2",
	"v0.36.0":     "v0.41.0",
	"v0.35.1":     "v0.40.0",
	"v0.35.0":     "v0.40.0",
	"v0.34.0":     "v0.39.0",
	"v0.33.0":     "v0.38.0",
	"v0.32.1":     "v0.37.0",
	"v0.31.0":     "v0.36.0",
	"v0.30.0":     "v0.35.0",
	"v0.29.0":     "v0.34.0",
	"v0.28.0":     "v0.33.0",
	"v0.27.0":     "v0.33.0",
	"v0.26.1":     "v0.32.0",
	"v0.26.0":     "v0.32.0",
}

type FluxVersionInfo struct {
	FluxVersion             string
	SourceControllerVersion string
	FluxNamespace           string
}

// GetFluxVersion returns the Flux version that is used in the cluster.
func GetFluxVersion(ctx context.Context, log logger.Logger, kubeClient client.Client) (fluxVersionInfo *FluxVersionInfo, guessed bool, err error) {
	log.Actionf("Getting Flux version ...")

	namespaceList := v1.NamespaceList{}
	listOptions := client.MatchingLabels{
		coretypes.PartOfLabel: "flux",
	}

	var foundNamespace v1.Namespace

	if err := kubeClient.List(ctx, &namespaceList, listOptions); err != nil {
		log.Failuref("error getting the list of Flux objects")
		return nil, false, err
	} else {
		for _, item := range namespaceList.Items {
			if item.GetLabels()[coretypes.VersionLabel] != "" {
				foundNamespace = item
				break
			}
		}
	}

	if foundNamespace.GetName() == "" {
		// try hard-coded namespace
		if err := kubeClient.Get(ctx, client.ObjectKey{Name: "flux-system"}, &foundNamespace); err != nil {
			log.Failuref("error getting the flux-system namespace")
			return nil, false, err
		}
	}

	if foundNamespace.GetName() != "" {
		labels := foundNamespace.GetLabels()
		if labels == nil {
			return nil, false, fmt.Errorf("error getting Flux labels")
		}

		fluxVersion := labels[coretypes.VersionLabel]
		if fluxVersion != "" {
			// ok, we found the version
			return &FluxVersionInfo{
				FluxVersion:             fluxVersion,
				FluxNamespace:           foundNamespace.GetName(),
				SourceControllerVersion: "",
			}, false, nil
		}

		// Try to guess the version
		// 1. get the source-controller deployment object
		deployment := appsv1.Deployment{}
		if err := kubeClient.Get(ctx, client.ObjectKey{Name: "source-controller", Namespace: foundNamespace.GetName()}, &deployment); err != nil {
			log.Failuref("error getting the source-controller deployment")
			return nil, false, err
		}

		// 2. get the source-controller image version
		image := deployment.Spec.Template.Spec.Containers[0].Image
		if image == "" {
			return nil, false, fmt.Errorf("error getting the source-controller image")
		}

		// 3. get the source-controller version by parsing the image version, split at :
		//    e.g. ghcr.io/fluxcd/source-controller:v0.33.0
		sourceVersion := image[strings.LastIndex(image, ":")+1:]

		// 4. get the Flux version by looking up the source-controller version in the map
		fluxVersion = sourceVerToFluxVer[sourceVersion]
		if fluxVersion == "" {
			return nil, false, fmt.Errorf("error getting the Flux version")
		}

		return &FluxVersionInfo{
			FluxVersion:             fluxVersion,
			FluxNamespace:           foundNamespace.GetName(),
			SourceControllerVersion: sourceVersion,
		}, true, nil
	}

	return nil, false, fmt.Errorf("no flux version found")
}
