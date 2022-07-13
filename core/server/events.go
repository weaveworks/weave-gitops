package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListFluxEvents(ctx context.Context, msg *pb.ListFluxEventsRequest) (*pb.ListFluxEventsResponse, error) {
	k8s, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, doClientError(err)
	}

	if msg.InvolvedObject == nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad request: no object was specified")
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &corev1.EventList{}
	})

	kind := msg.InvolvedObject.Kind.String()
	kind = strings.TrimPrefix(kind, "Kind")

	fields := client.MatchingFields{
		"involvedObject.kind":      kind,
		"involvedObject.name":      msg.InvolvedObject.Name,
		"involvedObject.namespace": msg.InvolvedObject.Namespace,
	}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.InvolvedObject.Namespace, clist, fields); err != nil {
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
					Name:      e.ObjectMeta.Name,
					Reason:    e.Reason,
					Message:   e.Message,
					Timestamp: e.LastTimestamp.Format(time.RFC3339),
					Host:      e.Source.Host,
				})
			}
		}
	}

	return &pb.ListFluxEventsResponse{Events: events}, nil
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
		return status.Errorf(codes.PermissionDenied, err.Error())
	} else if k8serrors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
