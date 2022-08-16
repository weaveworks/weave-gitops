package types

import (
	"bytes"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func K8sObjectToProto(object client.Object, clusterName string, tenant string) (*pb.Object, error) {
	var buf bytes.Buffer

	serializer := json.NewSerializer(json.DefaultMetaFactory, nil, nil, false)
	if err := serializer.Encode(object, &buf); err != nil {
		return nil, err
	}

	obj := &pb.Object{
		Payload:     buf.String(),
		ClusterName: clusterName,
		Tenant:		 tenant,
	}

	return obj, nil
}
