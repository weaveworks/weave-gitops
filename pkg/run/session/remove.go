package session

import (
	"context"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/hashicorp/go-multierror"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type InternalSession struct {
	SessionName      string
	SessionNamespace string
	PortForward      []string
	CliVersion       string
	Command          string
	Namespace        string
}

func Remove(kubeClient client.Client, session *InternalSession) error {
	var (
		helmRelease helmv2.HelmRelease
		result      error
	)

	if err := kubeClient.Get(context.Background(),
		types.NamespacedName{
			Namespace: session.SessionNamespace,
			Name:      session.SessionName,
		}, &helmRelease); err != nil {
		result = multierror.Append(result, err)
	}

	if err := kubeClient.Delete(context.Background(), &helmRelease); err != nil {
		result = multierror.Append(result, err)
	}

	timeout := 5 * time.Minute
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
	defer timeoutCancel()

	if err := wait.PollUntilContextTimeout(timeoutCtx, 2*time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		instance := appsv1.StatefulSet{}
		if err := kubeClient.Get(
			ctx,
			types.NamespacedName{
				Namespace: session.SessionNamespace,
				Name:      session.SessionName},
			&instance,
		); err != nil && apierrors.IsNotFound(err) {
			return true, nil
		} else if err != nil {
			return false, err
		}

		return false, nil
	}); err != nil {
		result = multierror.Append(result, err)
	}

	defer timeoutCancel()

	if err := wait.PollUntilContextTimeout(timeoutCtx, 2*time.Second, timeout, false, func(ctx context.Context) (bool, error) {
		pvc := corev1.PersistentVolumeClaim{}
		if err := kubeClient.Get(
			context.Background(),
			types.NamespacedName{
				Namespace: session.SessionNamespace,
				Name:      fmt.Sprintf("data-%s-0", session.SessionName),
			},
			&pvc,
		); err != nil && apierrors.IsNotFound(err) {
			return true, nil
		} else if err != nil {
			return false, err
		}

		if err := kubeClient.Delete(context.Background(), &pvc); err != nil {
			return false, err
		}
		return false, nil
	}); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}
