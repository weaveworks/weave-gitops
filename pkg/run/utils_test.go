package run

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
)

type isLocalClusterTest struct {
	clusterName string
	expected    bool
}

func runIsLocalClusterTest(client *kube.KubeHTTP, test isLocalClusterTest) {
	actual := IsLocalCluster(client)

	Expect(actual).To(Equal(test.expected))
}

var _ = Describe("IsLocalCluster", func() {
	var fakeKube *kube.KubeHTTP

	BeforeEach(func() {
		fakeKube = &kube.KubeHTTP{}
	})

	It("returns true for kind prefix", func() {
		test := isLocalClusterTest{
			clusterName: "kind-wego-dev",
			expected:    true,
		}

		fakeKube.ClusterName = test.clusterName

		runIsLocalClusterTest(fakeKube, test)
	})

	It("returns true for k3d prefix", func() {
		test := isLocalClusterTest{
			clusterName: "k3d-wego-dev",
			expected:    true,
		}

		fakeKube.ClusterName = test.clusterName

		runIsLocalClusterTest(fakeKube, test)
	})

	It("returns true if cluster name is minikube", func() {
		test := isLocalClusterTest{
			clusterName: "minikube",
			expected:    true,
		}

		fakeKube.ClusterName = test.clusterName

		runIsLocalClusterTest(fakeKube, test)
	})

	It("returns true if cluster name is docker-for-desktop", func() {
		test := isLocalClusterTest{
			clusterName: "docker-for-desktop",
			expected:    true,
		}

		fakeKube.ClusterName = test.clusterName

		runIsLocalClusterTest(fakeKube, test)
	})

	It("returns true if cluster name is docker-desktop", func() {
		test := isLocalClusterTest{
			clusterName: "docker-desktop",
			expected:    true,
		}

		fakeKube.ClusterName = test.clusterName

		runIsLocalClusterTest(fakeKube, test)
	})

	It("returns false for a gke cluster", func() {
		test := isLocalClusterTest{
			clusterName: "gke_testing_cluster-1",
			expected:    false,
		}

		fakeKube.ClusterName = test.clusterName

		runIsLocalClusterTest(fakeKube, test)
	})

	It("returns false for an empty string", func() {
		test := isLocalClusterTest{
			clusterName: "",
			expected:    false,
		}

		fakeKube.ClusterName = test.clusterName

		runIsLocalClusterTest(fakeKube, test)
	})
})

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
