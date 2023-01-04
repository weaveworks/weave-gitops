package telemetry

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"golang.org/x/crypto/sha3"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func InitTelemetry(ctx context.Context, cl cluster.Cluster) error {
	serverClient, err := cl.GetServerClient()
	if err != nil {
		return fmt.Errorf("failed to get server client; %w", err)
	}

	ns := &v1.Namespace{}

	err = serverClient.Get(ctx, client.ObjectKey{Name: "kube-system"}, ns)
	if err != nil {
		return fmt.Errorf("failed to get cluster namespace; %w", err)
	}

	key := []byte("VyzGoWoKvtJHyTnU+GVhDe+wU9bwZDH87bp505/0f/2UIpHzB+tmyZmfsH8/iJoH")
	buf := []byte(ns.GetUID())
	h := make([]byte, 32)
	d := sha3.NewShake128()

	_, err = d.Write(key)
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

	return nil
}
