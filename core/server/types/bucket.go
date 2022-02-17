package types

import (
	"strings"
	"time"

	"github.com/fluxcd/pkg/apis/meta"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToBucket(bucket *pb.Bucket) v1beta1.Bucket {
	labels := getGitopsLabelMap(bucket.Name)

	return v1beta1.Bucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta1.BucketKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bucket.Name,
			Namespace: bucket.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1.BucketSpec{
			BucketName: bucket.Name,
			Endpoint:   bucket.Endpoint,
			Insecure:   bucket.Insecure,
			Provider:   strings.ToLower(bucket.Provider.String()),
			Region:     bucket.Region,
			Interval:   metav1.Duration{Duration: time.Minute * 1},
			SecretRef: &meta.LocalObjectReference{
				Name: bucket.SecretRefName,
			},
			Timeout: &metav1.Duration{Duration: time.Second * 20},
		},
		Status: v1beta1.BucketStatus{},
	}
}

func BucketToProto(bucket *v1beta1.Bucket) *pb.Bucket {
	var provider pb.Bucket_Provider

	switch bucket.Spec.Provider {
	case v1beta1.GenericBucketProvider:
		provider = pb.Bucket_Generic
	case v1beta1.AmazonBucketProvider:
		provider = pb.Bucket_AWS
	case v1beta1.GoogleBucketProvider:
		provider = pb.Bucket_GCP
	}

	hr := &pb.Bucket{
		Name:      bucket.Name,
		Namespace: bucket.Namespace,
		Endpoint:  bucket.Spec.Endpoint,
		Insecure:  bucket.Spec.Insecure,
		Provider:  provider,
		Region:    bucket.Spec.Region,
		Interval: &pb.Interval{
			Minutes: 1,
		},
		SecretRefName: bucket.Spec.SecretRef.Name,
		Conditions:    mapConditions(bucket.Status.Conditions),
	}

	return hr
}
