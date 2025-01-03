package types

import (
	"bytes"

	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

type HelmReleaseStorage struct {
	Name     string `json:"name,omitempty"`
	Manifest string `json:"manifest,omitempty"`
}

func K8sObjectToProto(object client.Object, clusterName, tenant string, inventory []*pb.GroupVersionKind, info string) (*pb.Object, error) {
	var buf bytes.Buffer

	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{})
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
