package automation

import (
	"crypto/md5"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/models"
)

type ClusterAutomation struct {
	AppCRD                      models.Manifest
	GitOpsRuntime               models.Manifest
	SourceManifest              models.Manifest
	SystemKustomizationManifest models.Manifest
	SystemKustResourceManifest  models.Manifest
	UserKustResourceManifest    models.Manifest
	WegoAppManifest             models.Manifest

	// Do I really need this here? This is not needed for automation itself
	// Maybe we need to change the name to something like ClusterManifests
	WegoConfigManifest models.Manifest
}

func GetClusterHash(c models.Cluster) string {
	return fmt.Sprintf("wego-%x", md5.Sum([]byte(c.Name)))
}

func workAroundFluxDroppingDot(str string) string {
	return "." + str
}
