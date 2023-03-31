package health

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHealthCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	u := unstructured.Unstructured{}
	s := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	sData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
	g.Expect(err).To(BeNil())

	fmt.Println(sData)
	u.SetUnstructuredContent(sData)

	hc := NewHealthChecker()

	hs, err := hc.Check(u)
	g.Expect(err).To(BeNil())

	g.Expect(hs.Status).To(Equal(HealthStatusHealthy))
}
