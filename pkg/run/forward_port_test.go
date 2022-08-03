package run_test

import (
	"context"
	"errors"
	"github.com/weaveworks/weave-gitops/pkg/run"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// mock controller-runtime client
type mockClientForGetPodFromSpecMap struct {
	client.Client
	state stateGetPodFromSpecMap
}

type stateGetPodFromSpecMap string

const (
	stateListReturnErr    stateGetPodFromSpecMap = "list-return-err"
	stateListNoRunningPod stateGetPodFromSpecMap = "list-no-running-pod"
	stateListZeroPod      stateGetPodFromSpecMap = "list-zero-pod"
	stateListHasPod       stateGetPodFromSpecMap = "list-has-pod"

	stateGetReturnErr stateGetPodFromSpecMap = "get-return-err"
)

func (c *mockClientForGetPodFromSpecMap) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	switch c.state {
	case stateListReturnErr:
		return errors.New("fake error")

	default:
		listOptions := &client.ListOptions{}
		for _, opt := range opts {
			opt.ApplyToList(listOptions)
		}

		podList := &corev1.PodList{}

		switch c.state {
		case stateListZeroPod:
			podList.Items = []corev1.Pod{}

		case stateListNoRunningPod:
			podList.Items = append(podList.Items, corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: listOptions.Namespace,
				},
			})

		case stateListHasPod:
			podList.Items = append(podList.Items, corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: listOptions.Namespace,
				},
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{},
					Phase:      corev1.PodRunning,
				},
			})
		}

		podList.DeepCopyInto(list.(*corev1.PodList))
	}

	return nil
}

func (c *mockClientForGetPodFromSpecMap) Get(_ context.Context, key client.ObjectKey, obj client.Object) error {
	switch c.state {
	case stateGetReturnErr:
		return errors.New("fake error")

	default:
		switch obj := obj.(type) {
		case *corev1.Pod:
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
			}
			pod.DeepCopyInto(obj)
		case *corev1.Service:
			service := corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": key.Name,
					},
				},
			}
			service.DeepCopyInto(obj)
		case *appsv1.Deployment:
			deployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": key.Name,
						},
					},
				},
			}
			deployment.DeepCopyInto(obj)
		}
	}

	return nil
}

var _ = Describe("GetPodFromSpecMap", func() {
	It("should return an error if the pod spec is not correct", func() {
		_, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Kind: "something",
		}, &mockClientForGetPodFromSpecMap{})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unsupported spec kind"))
	})

	It("should return an error if the client returns an error", func() {
		_, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "pod",
		}, &mockClientForGetPodFromSpecMap{state: stateGetReturnErr})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake error"))
	})

	It("returns a pod according to the pod spec", func() {
		pod, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "pod",
		}, &mockClientForGetPodFromSpecMap{})

		Expect(err).To(BeNil())
		Expect(pod.Name).To(Equal("name"))
		Expect(pod.Namespace).To(Equal("ns"))
	})

	// Service tests

	It("should return an error if the client returns an error", func() {
		_, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "service",
		}, &mockClientForGetPodFromSpecMap{state: stateGetReturnErr})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("error getting service: fake error, namespaced Name: ns/name"))
	})

	It("should return an error if the client returns an error", func() {
		_, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "service",
		}, &mockClientForGetPodFromSpecMap{state: stateListReturnErr})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake error"))
	})

	It("returns a pod according to the service spec", func() {
		pod, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "service",
		}, &mockClientForGetPodFromSpecMap{state: stateListHasPod})

		Expect(err).To(BeNil())
		Expect(pod.Name).To(Equal("pod-1"))
		Expect(pod.Namespace).To(Equal("ns"))
	})

	It("returns a pod according to the service spec", func() {
		pod, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "service",
		}, &mockClientForGetPodFromSpecMap{state: stateListZeroPod})

		Expect(err).To(HaveOccurred())
		Expect(pod).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("no pods found for service"))
	})

	It("returns a pod according to the service spec", func() {
		pod, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "service",
		}, &mockClientForGetPodFromSpecMap{state: stateListNoRunningPod})

		Expect(err).To(HaveOccurred())
		Expect(pod).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("no running pods found for service"))
	})

	// Deployment tests

	It("should return an error if the client returns an error", func() {
		_, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "deployment",
		}, &mockClientForGetPodFromSpecMap{state: stateGetReturnErr})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("error getting deployment: fake error, namespaced Name: ns/name"))
	})

	It("should return an error if the client returns an error", func() {
		_, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "deployment",
		}, &mockClientForGetPodFromSpecMap{state: stateListReturnErr})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("fake error"))
	})

	It("returns a pod according to the deployment spec", func() {
		pod, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "deployment",
		}, &mockClientForGetPodFromSpecMap{state: stateListHasPod})

		Expect(err).To(BeNil())
		Expect(pod.Name).To(Equal("pod-1"))
		Expect(pod.Namespace).To(Equal("ns"))
	})

	It("returns a pod according to the deployment spec", func() {
		pod, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "deployment",
		}, &mockClientForGetPodFromSpecMap{state: stateListZeroPod})

		Expect(err).To(HaveOccurred())
		Expect(pod).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("no pods found for deployment"))
	})

	It("returns a pod according to the deployment spec", func() {
		pod, err := run.GetPodFromSpecMap(&run.PortForwardSpec{
			Namespace: "ns",
			Name:      "name",
			Kind:      "deployment",
		}, &mockClientForGetPodFromSpecMap{state: stateListNoRunningPod})

		Expect(err).To(HaveOccurred())
		Expect(pod).To(BeNil())
		Expect(err.Error()).To(ContainSubstring("no running pods found for deployment"))
	})
})
