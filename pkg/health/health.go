package health

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
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
	case "DaemonSet":
		return checkDaemonSet(obj)
	case "StatefulSet":
		return checkStatefulSet(obj)
	case "Job":
		return checkJob(obj)
	case "Pod":
		return checkPod(obj)
	}

	return HealthStatus{
		Status: HealthStatusUnknown,
	}, nil
}

func checkDeployment(obj unstructured.Unstructured) (HealthStatus, error) {
	var dpl appsv1.Deployment

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &dpl)
	if err != nil {
		err = fmt.Errorf("converting unstructured to deployment: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	if dpl.Generation != dpl.Status.ObservedGeneration {
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
	var rs appsv1.ReplicaSet

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &rs)
	if err != nil {
		err = fmt.Errorf("converting unstructured to replicaset: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	cond := getReplicaSetCondition(rs.Status, appsv1.ReplicaSetReplicaFailure)
	if cond != nil && cond.Status == corev1.ConditionTrue {
		return HealthStatus{Status: HealthStatusUnhealthy, Message: cond.Message}, nil
	}

	if rs.Generation != rs.Status.ObservedGeneration {
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

func checkDaemonSet(obj unstructured.Unstructured) (HealthStatus, error) {
	var ds appsv1.DaemonSet

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &ds)
	if err != nil {
		err = fmt.Errorf("converting unstructured to daemonset: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	if ds.Generation != ds.Status.ObservedGeneration {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting spec to be observed"}, nil
	}

	if ds.Status.UpdatedNumberScheduled != ds.Status.DesiredNumberScheduled {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting updated number scheduled to be equal to desired number scheduled"}, nil
	}

	if ds.Status.NumberAvailable != ds.Status.DesiredNumberScheduled {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting for available number to be equal to desired number scheduled"}, nil
	}

	return HealthStatus{Status: HealthStatusHealthy}, nil
}

func checkStatefulSet(obj unstructured.Unstructured) (HealthStatus, error) {
	var sts appsv1.StatefulSet

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &sts)
	if err != nil {
		err = fmt.Errorf("converting unstructured to statefulset: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	if sts.Generation != sts.Status.ObservedGeneration {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting spec to be observed"}, nil
	}

	if sts.Spec.Replicas != nil && *sts.Spec.Replicas != sts.Status.ReadyReplicas {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting for ready replicas"}, nil
	}

	//ref: https://github.com/kubernetes/kubernetes/blob/5232ad4a00ec93942d0b2c6359ee6cd1201b46bc/pkg/kubectl/rollout_status.go#L137
	if sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
		if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
				return HealthStatus{
					Status: HealthStatusProgressing,
					Message: fmt.Sprintf("Waiting for partitioned roll out to finish: %d out of %d new pods have been updated...\n",
						sts.Status.UpdatedReplicas, (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition))}, nil
			}
		}

		return HealthStatus{
			Status: HealthStatusHealthy,
		}, nil
	}

	if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
		return HealthStatus{
			Status: HealthStatusProgressing,
			Message: fmt.Sprintf("waiting for statefulset rolling update to complete %d pods at revision %s...\n",
				sts.Status.UpdatedReplicas, sts.Status.UpdateRevision),
		}, nil
	}

	if sts.Spec.Replicas != nil && *sts.Spec.Replicas != sts.Status.ReadyReplicas {
		return HealthStatus{Status: HealthStatusProgressing, Message: "waiting for ready replicas"}, nil
	}

	return HealthStatus{Status: HealthStatusHealthy}, nil
}

func checkJob(obj unstructured.Unstructured) (HealthStatus, error) {
	var job batchv1.Job

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &job)
	if err != nil {
		err = fmt.Errorf("converting unstructured to job: %w", err)
		return HealthStatus{Status: HealthStatusUnknown, Message: err.Error()}, err
	}

	if job.Status.Succeeded > 0 {
		return HealthStatus{Status: HealthStatusHealthy}, nil
	}

	if job.Status.Failed > 0 {
		return HealthStatus{Status: HealthStatusUnhealthy, Message: "job is in a failed state."}, nil
	}

	return HealthStatus{Status: HealthStatusProgressing}, nil
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
