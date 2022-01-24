package server

import (
	"context"

	"github.com/fluxcd/source-controller/api/v1beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func listGitRepostories(ctx context.Context, k8s client.Client, namespace string, opts client.MatchingLabels) ([]v1beta1.GitRepository, error) {
	list := &sourcev1.GitRepositoryList{}

	if err := k8s.List(ctx, list, opts); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get git repository list: %s", err.Error())
	}

	return list.Items, nil
}
