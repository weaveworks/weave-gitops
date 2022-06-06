package types

import (
	"bytes"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// This method whitelists legal object kinds to request
func GetGVK(kind string) (*schema.GroupVersionKind, error) {
	var res schema.GroupVersionKind

	switch kind {
	case kustomizev1.KustomizationKind:
		res = kustomizev1.GroupVersion.WithKind(kustomizev1.KustomizationKind)
	case helmv2.HelmReleaseKind:
		res = helmv2.GroupVersion.WithKind(helmv2.HelmReleaseKind)
	case sourcev1.GitRepositoryKind:
		res = sourcev1.GroupVersion.WithKind(sourcev1.GitRepositoryKind)
	case sourcev1.HelmChartKind:
		res = sourcev1.GroupVersion.WithKind(sourcev1.HelmChartKind)
	case sourcev1.HelmRepositoryKind:
		res = sourcev1.GroupVersion.WithKind(sourcev1.HelmRepositoryKind)
	case sourcev1.BucketKind:
		res = sourcev1.GroupVersion.WithKind(sourcev1.BucketKind)
	default:
		return nil, fmt.Errorf("Looking up objects of type %v not supported", kind)
	}

	return &res, nil
}

func K8sObjectToProto(object client.Object, clusterName string) (*pb.Object, error) {
	var buf bytes.Buffer

	serializer := json.NewSerializer(json.DefaultMetaFactory, nil, nil, false)
	if err := serializer.Encode(object, &buf); err != nil {
		return nil, err
	}

	obj := &pb.Object{
		Payload:     buf.String(),
		ClusterName: clusterName,
	}

	return obj, nil
}
