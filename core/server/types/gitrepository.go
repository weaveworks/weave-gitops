package types

import (
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	managedByWeaveGitops      = "weave-gitops"
	createdBySourceController = "source-controller"
)

func ProtoToGitRepository(repo *pb.GitRepository) *v1beta1.GitRepository {
	labels := getGitopsLabelMap(repo.Name)

	return &v1beta1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta2.KustomizationKind,
			APIVersion: v1beta2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      repo.Name,
			Namespace: repo.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1.GitRepositorySpec{
			URL: repo.Url,
			//Interval: intervalDuration(kustomization.Interval),
			SecretRef: &meta.LocalObjectReference{
				Name: repo.SecretRef,
			},
			Reference: &v1beta1.GitRepositoryRef{
				Branch: repo.Reference.Branch,
				Tag:    repo.Reference.Tag,
				SemVer: repo.Reference.Semver,
				Commit: repo.Reference.Commit,
			},
		},
		Status: v1beta1.GitRepositoryStatus{},
	}
}

func GitRepositoryToProto(repository *v1beta1.GitRepository) *pb.GitRepository {
	gr := &pb.GitRepository{
		Name:      repository.Name,
		Namespace: repository.Namespace,
		Url:       repository.Spec.URL,
		Reference: &pb.GitRepositoryRef{
			Branch: repository.Spec.Reference.Branch,
			Tag:    repository.Spec.Reference.Tag,
			Semver: repository.Spec.Reference.SemVer,
			Commit: repository.Spec.Reference.Commit,
		},
		Interval: &pb.Interval{
			Hours:   int64(repository.Spec.Interval.Hours()),
			Minutes: int64(repository.Spec.Interval.Minutes()) % 60,
			Seconds: int64(repository.Spec.Interval.Seconds()) % 60,
		},
		Conditions: mapConditions(repository.Status.Conditions),
		Suspended:  repository.Spec.Suspend,
	}

	if repository.Spec.SecretRef != nil {
		gr.SecretRef = repository.Spec.SecretRef.Name
	}

	return gr
}
