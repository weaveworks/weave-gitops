package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
)

const (
	fluxNamespacePartOf   = "flux"
	fluxNamespaceInstance = "flux-system"
)

var ErrNamespaceNotFound = errors.New("namespace not found")

func (as *coreServer) GetFluxNamespace(ctx context.Context, msg *pb.GetFluxNamespaceRequest) (*pb.GetFluxNamespaceResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	nsList := corev1.NamespaceList{}
	options := matchLabel(
		withPartOfLabel(fluxNamespacePartOf),
		withInstanceLabel(fluxNamespaceInstance),
	)

	if err = k8s.List(ctx, &nsList, &options); err != nil {
		return nil, doClientError(err)
	}

	if len(nsList.Items) == 0 {
		return nil, ErrNamespaceNotFound
	}

	return &pb.GetFluxNamespaceResponse{Name: nsList.Items[0].Name}, nil
}

func (as *coreServer) ListNamespaces(ctx context.Context, msg *pb.ListNamespacesRequest) (*pb.ListNamespacesResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	nsList := corev1.NamespaceList{}

	if err = k8s.List(ctx, &nsList); err != nil {
		return nil, doClientError(err)
	}

	auth, err := newAuthClient(as.rest)
	if err != nil {
		return nil, fmt.Errorf("making auth client: %w", err)
	}

	response := &pb.ListNamespacesResponse{
		Namespaces: []*pb.Namespace{},
	}

	for _, ns := range nsList.Items {
		sar := &authorizationv1.SelfSubjectRulesReview{
			Spec: authorizationv1.SelfSubjectRulesReviewSpec{
				Namespace: ns.Name,
			},
		}

		authRes, err := auth.SelfSubjectRulesReviews().Create(ctx, sar, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}

		response.Namespaces = append(response.Namespaces, types.NamespaceToProto(ns, authRes))
	}

	return response, nil
}

func newAuthClient(cfg *rest.Config) (typedauth.AuthorizationV1Interface, error) {
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("making clientset: %w", err)
	}

	return cs.AuthorizationV1(), nil
}
