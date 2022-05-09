package server

import (
	"context"
	"fmt"
	"time"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListFluxEvents(ctx context.Context, msg *pb.ListFluxEventsRequest) (*pb.ListFluxEventsResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	if msg.InvolvedObject == nil {
		return nil, status.Errorf(codes.InvalidArgument, "bad request: no object was specified")
	}

	l := &corev1.EventList{}

	fields := client.MatchingFields{
		"involvedObject.kind":      msg.InvolvedObject.Kind,
		"involvedObject.name":      msg.InvolvedObject.Name,
		"involvedObject.namespace": msg.InvolvedObject.Namespace,
	}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.InvolvedObject.Namespace, l, fields); err != nil {
		return nil, fmt.Errorf("could not get events: %w", err)
	}

	events := []*pb.Event{}

	for _, e := range l.Items {
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

	return &pb.ListFluxEventsResponse{Events: events}, nil
}
