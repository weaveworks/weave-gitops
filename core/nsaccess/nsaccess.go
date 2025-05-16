package nsaccess

import (
	"context"
	"fmt"

	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedauth "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Checker contains methods for validing user access to Kubernetes namespaces, based on a set of PolicyRules
//
//counterfeiter:generate . Checker
type Checker interface {
	// FilterAccessibleNamespaces returns a filtered list of namespaces to which a user has access to
	FilterAccessibleNamespaces(ctx context.Context, auth typedauth.AuthorizationV1Interface, namespaces []corev1.Namespace) ([]corev1.Namespace, error)
}

type simpleChecker struct{}

func NewChecker() Checker {
	return simpleChecker{}
}

func (sc simpleChecker) FilterAccessibleNamespaces(ctx context.Context, auth typedauth.AuthorizationV1Interface, namespaces []corev1.Namespace) ([]corev1.Namespace, error) {
	accessToAllNamespace, err := hasAccessToAllNamespaces(ctx, auth)
	if err != nil {
		return nil, err
	}

	if accessToAllNamespace {
		return namespaces, nil
	}

	var result []corev1.Namespace
	for _, ns := range namespaces {
		ok, err := hasAccessToNamespace(ctx, auth, ns)
		if err != nil {
			return nil, fmt.Errorf("user namespace access: %w", err)
		}

		if ok {
			result = append(result, ns)
		}
	}

	return result, nil
}

func hasAccessToNamespace(ctx context.Context, auth typedauth.AuthorizationV1Interface, ns corev1.Namespace) (bool, error) {
	ssar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:      "get",
				Resource:  "configmaps",
				Namespace: ns.Name,
			},
		},
	}
	res, err := auth.SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return res.Status.Allowed, nil
}

func hasAccessToAllNamespaces(ctx context.Context, auth typedauth.AuthorizationV1Interface) (bool, error) {
	ssar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:     "list",
				Resource: "namespaces",
			},
		},
	}
	res, err := auth.SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return res.Status.Allowed, nil
}
