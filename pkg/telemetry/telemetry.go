package telemetry

import (
	"encoding/hex"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"golang.org/x/crypto/sha3"
	"k8s.io/apimachinery/pkg/types"
)

func InitTelemetry(factory clustersmngr.ClustersManager) error {
	if featureflags.Get("WEAVE_GITOPS_FEATURE_TELEMETRY") == "true" {
		var namespace types.UID

		namespaces := factory.GetClustersNamespaces()["Default"]
		for _, ns := range namespaces {
			if ns.GetName() == "kube-system" {
				namespace = ns.GetUID()
			}
		}

		key := []byte("VyzGoWoKvtJHyTnU+GVhDe+wU9bwZDH87bp505/0f/2UIpHzB+tmyZmfsH8/iJoH")
		buf := []byte(namespace)
		h := make([]byte, 32)
		d := sha3.NewShake128()

		_, err := d.Write(key)
		if err != nil {
			return err
		}

		_, err = d.Write(buf)
		if err != nil {
			return err
		}

		_, err = d.Read(h)
		if err != nil {
			return err
		}

		featureflags.Set("ACCOUNT_ID", hex.EncodeToString(h))
	}

	return nil
}
