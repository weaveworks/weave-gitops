package types

import (
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/gitops-server/pkg/api/core"
)

func BucketToProto(bucket *v1beta1.Bucket, clusterName string) *pb.Bucket {
	var provider pb.Bucket_Provider

	switch bucket.Spec.Provider {
	case v1beta1.GenericBucketProvider:
		provider = pb.Bucket_Generic
	case v1beta1.AmazonBucketProvider:
		provider = pb.Bucket_AWS
	case v1beta1.GoogleBucketProvider:
		provider = pb.Bucket_GCP
	}

	bkt := &pb.Bucket{
		Name:      bucket.Name,
		Namespace: bucket.Namespace,
		Endpoint:  bucket.Spec.Endpoint,
		Insecure:  bucket.Spec.Insecure,
		Provider:  provider,
		Region:    bucket.Spec.Region,
		Interval:  durationToInterval(bucket.Spec.Interval),
		// SecretRefName: bucket.Spec.SecretRef.Name,
		Conditions:    mapConditions(bucket.Status.Conditions),
		Suspended:     bucket.Spec.Suspend,
		BucketName:    bucket.Spec.BucketName,
		LastUpdatedAt: lastUpdatedAt(bucket),
		ClusterName:   clusterName,
	}

	if bucket.Spec.SecretRef != nil {
		bkt.SecretRefName = bucket.Spec.SecretRef.Name
	}

	return bkt
}
