package install

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

type isPodStatusConditionPresentAndEqualTest struct {
	conditions      []corev1.PodCondition
	conditionType   corev1.PodConditionType
	conditionStatus corev1.ConditionStatus
	expected        bool
}

func runIsPodStatusConditionPresentAndEqualTest(test isPodStatusConditionPresentAndEqualTest) {
	actual := isPodStatusConditionPresentAndEqual(test.conditions, test.conditionType, test.conditionStatus)

	Expect(actual).To(Equal(test.expected))
}

var _ = Describe("isPodStatusConditionPresentAndEqual", func() {
	It("returns true if condition statuses are the same and condition is true", func() {
		test := isPodStatusConditionPresentAndEqualTest{
			conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			conditionType:   corev1.PodReady,
			conditionStatus: corev1.ConditionTrue,
			expected:        true,
		}

		runIsPodStatusConditionPresentAndEqualTest(test)
	})

	It("returns true if condition statuses are the same and condition is false", func() {
		test := isPodStatusConditionPresentAndEqualTest{
			conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionFalse},
			},
			conditionType:   corev1.PodReady,
			conditionStatus: corev1.ConditionFalse,
			expected:        true,
		}
		runIsPodStatusConditionPresentAndEqualTest(test)
	})

	It("returns false if condition statuses are different", func() {
		test := isPodStatusConditionPresentAndEqualTest{
			conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionUnknown},
			},
			conditionType:   corev1.PodReady,
			conditionStatus: corev1.ConditionTrue,
			expected:        false,
		}

		runIsPodStatusConditionPresentAndEqualTest(test)
	})

	It("returns false if condition types are different and condition statuses are the same", func() {
		test := isPodStatusConditionPresentAndEqualTest{
			conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			conditionType:   corev1.PodInitialized,
			conditionStatus: corev1.ConditionTrue,
			expected:        false,
		}

		runIsPodStatusConditionPresentAndEqualTest(test)
	})
})
