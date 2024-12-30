package server

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/fluxsync"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

const (
	SuspendedByAnnotation      = "metadata.weave.works/suspended-by"
	SuspendedCommentAnnotation = "metadata.weave.works/suspended-comment"
)

func (cs *coreServer) ToggleSuspendResource(ctx context.Context, msg *pb.ToggleSuspendResourceRequest) (*pb.ToggleSuspendResourceResponse, error) {
	principal := auth.Principal(ctx)
	respErrors := multierror.Error{}

	for _, obj := range msg.Objects {
		clusterName := obj.ClusterName
		clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, principal)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error getting impersonating client: %w", err), respErrors.Errors...)
			continue
		}

		c, err := clustersClient.Scoped(clusterName)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("getting cluster client: %w", err), respErrors.Errors...)
			continue
		}

		key := client.ObjectKey{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		}

		gvk, err := cs.primaryKinds.Lookup(obj.Kind)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("looking up GVK for %q: %w", obj.Kind, err), respErrors.Errors...)
			continue
		}

		obj := fluxsync.ToReconcileable(*gvk)

		log := cs.logger.WithValues(
			"user", principal.ID,
			"kind", obj.GroupVersionKind().Kind,
			"name", key.Name,
			"namespace", key.Namespace,
			"principal", principal.ID,
			"cluster", clusterName,
		)

		if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("getting reconcilable object: %w", err), respErrors.Errors...)
			continue
		}

		patch := client.MergeFrom(obj.DeepCopyClientObject())

		err = obj.SetSuspended(msg.Suspend)
		if err != nil {
			return nil, err
		}

		changeSuspendAnnotations(obj, msg.Suspend, msg.Comment, principal)

		if msg.Suspend {
			log.Info("Suspending resource")
		} else {
			log.Info("Resuming resource")
		}

		if err := c.Patch(ctx, obj.AsClientObject(), patch); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("patching object: %w", err), respErrors.Errors...)
		}
	}

	return &pb.ToggleSuspendResourceResponse{}, respErrors.ErrorOrNil()
}

func changeSuspendAnnotations(obj fluxsync.Reconcilable, suspend bool, comment string, principal *auth.UserPrincipal) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	if suspend {
		annotations[SuspendedByAnnotation] = principal.ID
		if comment != "" {
			annotations[SuspendedCommentAnnotation] = comment
		}
		obj.SetAnnotations(annotations)
	} else {
		delete(annotations, SuspendedByAnnotation)
		delete(annotations, SuspendedCommentAnnotation)
		obj.SetAnnotations(annotations)
	}
}
