package install

import (
	corev1 "k8s.io/api/core/v1"
)

// isPodStatusConditionPresentAndEqual returns true when conditionType is present and equal to status.
func isPodStatusConditionPresentAndEqual(conditions []corev1.PodCondition, conditionType corev1.PodConditionType, status corev1.ConditionStatus) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}

	return false
}
