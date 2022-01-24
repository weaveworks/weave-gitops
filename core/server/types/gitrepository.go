package types

import (
	"github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	managedByWeaveGitops         = "weave-gitops"
	createdByKustomizeController = "kustomize-controller"
	createdBySourceController    = "source-controller"
)

func ProtoToGitRepository(repositoryReq *pb.AddGitRepositoryReq) *v1beta1.GitRepository {
	labels := map[string]string{
		"app.kubernetes.io/managed-by": managedByWeaveGitops,
		"app.kubernetes.io/created-by": createdBySourceController,
	}

	if repositoryReq.AppName != "" {
		labels["app.kubernetes.io/part-of"] = repositoryReq.AppName
	}

	return &v1beta1.GitRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta2.KustomizationKind,
			APIVersion: v1beta2.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      repositoryReq.Name,
			Namespace: repositoryReq.Namespace,
			Labels:    labels,
		},
		Spec: v1beta1.GitRepositorySpec{
			URL: repositoryReq.Url,
			//Interval: intervalDuration(kustomization.Interval),
			SecretRef: &meta.LocalObjectReference{
				Name: repositoryReq.SecretRef,
			},
			Reference: &v1beta1.GitRepositoryRef{
				Branch: repositoryReq.Reference.Branch,
				Tag:    repositoryReq.Reference.Tag,
				SemVer: repositoryReq.Reference.Semver,
				Commit: repositoryReq.Reference.Commit,
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
		Interval: nil,
	}

	if repository.Spec.SecretRef != nil {
		gr.SecretRef = repository.Spec.SecretRef.Name
	}

	return gr
}
