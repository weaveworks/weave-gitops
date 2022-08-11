package types

import (
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func BucketToProto(bucket *sourcev1.Bucket, clusterName string, tenant string) *pb.Bucket {
	var provider pb.Bucket_Provider

	switch bucket.Spec.Provider {
	case sourcev1.GenericBucketProvider:
		provider = pb.Bucket_Generic
	case sourcev1.AmazonBucketProvider:
		provider = pb.Bucket_AWS
	case sourcev1.GoogleBucketProvider:
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
		ApiVersion:    bucket.APIVersion,
		Tenant: 	   tenant,
	}

	if bucket.Spec.SecretRef != nil {
		bkt.SecretRefName = bucket.Spec.SecretRef.Name
	}

	return bkt
}
