package types

import (
	"bytes"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type HelmReleaseStorage struct {
	Name     string `json:"name,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

func K8sObjectToProto(object client.Object, clusterName, tenant string, inventory []*pb.GroupVersionKind, info string) (*pb.Object, error) {
	var buf bytes.Buffer

	serializer := json.NewSerializer(json.DefaultMetaFactory, nil, nil, false)
	if err := serializer.Encode(object, &buf); err != nil {
		return nil, err
	}

	obj := &pb.Object{
		Payload:     buf.String(),
		ClusterName: clusterName,
		Tenant:      tenant,
		Uid:         string(object.GetUID()),
		Inventory:   inventory,
		Info:        info,
	}

	return obj, nil
}
