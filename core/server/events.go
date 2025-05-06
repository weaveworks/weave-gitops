package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func (cs *coreServer) ListEvents(ctx context.Context, msg *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	if msg.InvolvedObject == nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad request: no object was specified")
	}

	var clustersClient clustersmngr.Client

	var err error

	if msg.InvolvedObject.ClusterName != "" {
		clustersClient, err = cs.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), msg.InvolvedObject.ClusterName)
	} else {
		clustersClient, err = cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	}

	if err != nil {
		return nil, doClientError(err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &corev1.EventList{}
	})

	kind := msg.InvolvedObject.Kind

	gvk, err := cs.primaryKinds.Lookup(kind)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad request: not a recognized object kind")
	}

	fields := client.MatchingFields{
		"involvedObject.kind":      gvk.Kind,
		"involvedObject.name":      msg.InvolvedObject.Name,
		"involvedObject.namespace": msg.InvolvedObject.Namespace,
	}

	if err := list(ctx, clustersClient, temporarilyEmptyAppName, msg.InvolvedObject.Namespace, clist, fields); err != nil {
		return nil, fmt.Errorf("could not get events: %w", err)
	}

	events := []*pb.Event{}

	for _, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*corev1.EventList)
			if !ok {
				continue
			}

			for _, e := range list.Items {
				events = append(events, &pb.Event{
					Type:      e.Type,
					Component: e.Source.Component,
					Name:      e.Name,
					Reason:    e.Reason,
					Message:   e.Message,
					Timestamp: e.LastTimestamp.Format(time.RFC3339),
					Host:      e.Source.Host,
				})
			}
		}
	}

	return &pb.ListEventsResponse{Events: events}, nil
}

func list(ctx context.Context, k8s clustersmngr.Client, appName, namespace string, list clustersmngr.ClusteredObjectList, extraOpts ...client.ListOption) error {
	opts := []client.ListOption{
		getMatchingLabels(appName),
		client.InNamespace(namespace),
	}

	opts = append(opts, extraOpts...)
	err := k8s.ClusteredList(ctx, list, true, opts...)
	err = wrapK8sAPIError("list resource", err)

	return err
}

func wrapK8sAPIError(msg string, err error) error {
	if k8serrors.IsUnauthorized(err) {
		return status.Errorf(codes.PermissionDenied, "%s", err.Error())
	} else if k8serrors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, "%s", err.Error())
	} else if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
