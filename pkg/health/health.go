package health

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Represents resource health status
type HealthStatusCode string

const (
	HealthStatusProgressing HealthStatusCode = "Progressing"
	HealthStatusHealthy     HealthStatusCode = "Healthy"
	HealthStatusUnhealthy   HealthStatusCode = "Unhealthy"
	HealthStatusUnknown     HealthStatusCode = "Unknown"
)

type HealthChecker interface {
	Check(obj unstructured.Unstructured) (HealthStatus, error)
}

type HealthStatus struct {
	Status  HealthStatusCode `json:"status,omitempty"`
	Message string           `json:"message,omitempty"`
}

func NewHealthChecker() HealthChecker {
	return &healthChecker{}
}

type healthChecker struct{}

func (hc *healthChecker) Check(obj unstructured.Unstructured) (HealthStatus, error) {
	gvk := obj.GroupVersionKind()

	switch gvk.Kind {
	case "Deployment":
		return checkDeployment(obj)
	case "ReplicaSet":
		return checkReplicaSet(obj)
	case "Pod":
		return checkPod(obj)
	}

	return HealthStatus{
		Status: HealthStatusUnknown,
	}, nil
}

func checkDeployment(obj unstructured.Unstructured) (HealthStatus, error) {
	var dpl v1.Deployment

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &dpl)
	if err != nil {
		err = fmt.Errorf("converting unstructured to deployment: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	if dpl.Generation < dpl.Status.ObservedGeneration {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting spec to be observed"}, nil
	}

	cond := getDeploymentCondition(dpl, appsv1.DeploymentProgressing)

	if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
		return HealthStatus{Status: HealthStatusUnhealthy, Message: "deployment exceeded its progress deadline"}, nil
	}

	if dpl.Spec.Replicas != nil && *dpl.Spec.Replicas != dpl.Status.UpdatedReplicas {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting for updated replicas"}, nil
	}

	return HealthStatus{Status: HealthStatusHealthy}, nil
}

func checkReplicaSet(obj unstructured.Unstructured) (HealthStatus, error) {
	var rs v1.ReplicaSet

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &rs)
	if err != nil {
		err = fmt.Errorf("converting unstructured to replicaset: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	cond := getReplicaSetCondition(rs.Status, appsv1.ReplicaSetReplicaFailure)
	if cond != nil && cond.Status == corev1.ConditionTrue {
		return HealthStatus{Status: HealthStatusUnhealthy, Message: cond.Message}, nil
	}

	if rs.Generation < rs.Status.ObservedGeneration {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting spec to be observed"}, nil
	}

	if rs.Spec.Replicas != nil && *rs.Spec.Replicas != rs.Status.AvailableReplicas {
		return HealthStatus{
			Status:  HealthStatusProgressing,
			Message: "waiting for replicas",
		}, nil
	}

	return HealthStatus{
		Status: HealthStatusHealthy,
	}, nil
}

func checkPod(obj unstructured.Unstructured) (HealthStatus, error) {
	var pod corev1.Pod

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &pod)
	if err != nil {
		err = fmt.Errorf("converting unstructured to pod: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	switch pod.Status.Phase {
	case corev1.PodPending:
		return HealthStatus{Status: HealthStatusProgressing, Message: pod.Status.Message}, nil
	case corev1.PodFailed:
		return HealthStatus{Status: HealthStatusUnhealthy, Message: pod.Status.Message}, nil
	case corev1.PodSucceeded, corev1.PodRunning:
		return HealthStatus{Status: HealthStatusHealthy, Message: pod.Status.Message}, nil
	}

	return HealthStatus{Status: HealthStatusUnknown, Message: pod.Status.Message}, nil
}

func getDeploymentCondition(deployment appsv1.Deployment, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range deployment.Status.Conditions {
		if deployment.Status.Conditions[i].Type == condType {
			return &deployment.Status.Conditions[i]
		}
	}
	return nil
}

func getReplicaSetCondition(status appsv1.ReplicaSetStatus, condType appsv1.ReplicaSetConditionType) *appsv1.ReplicaSetCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}
