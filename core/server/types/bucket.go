package types

import (
	"strings"
	"time"

	"github.com/fluxcd/pkg/apis/meta"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToBucket(bucketReq *pb.AddBucketReq) v1beta1.Bucket {
	labels := getGitopsLabelMap(bucketReq.AppName)

	return v1beta1.Bucket{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta1.BucketKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bucketReq.Bucket.Name,
			Namespace: bucketReq.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1.BucketSpec{
			BucketName: bucketReq.Bucket.Name,
			Endpoint:   bucketReq.Bucket.Endpoint,
			Insecure:   bucketReq.Bucket.Insecure,
			Provider:   strings.ToLower(bucketReq.Bucket.Provider.String()),
			Region:     bucketReq.Bucket.Region,
			Interval:   metav1.Duration{Duration: time.Minute * 1},
			SecretRef: &meta.LocalObjectReference{
				Name: bucketReq.Bucket.SecretRefName,
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
		Name:      bucket.Spec.BucketName,
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
