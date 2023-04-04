package health

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestHealthCheck(t *testing.T) {
	g := NewGomegaWithT(t)

	hc := NewHealthChecker()

	type scenarios struct {
		data         string
		healthStatus HealthStatusCode
	}

	for _, scenario := range []scenarios{
		{
			data:         "testdata/deployment-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/deployment-unhealthy.yaml",
			healthStatus: HealthStatusUnhealthy,
		},
		{
			data:         "testdata/deployment-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/replicaset-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/replicaset-unhealthy.yaml",
			healthStatus: HealthStatusUnhealthy,
		},
		{
			data:         "testdata/replicaset-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/pod-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/pod-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/pod-unhealthy.yaml",
			healthStatus: HealthStatusUnhealthy,
		},
	} {
		t.Run(fmt.Sprintf("%s is %s", scenario.data, scenario.healthStatus), func(t *testing.T) {
			yamlBytes, err := os.ReadFile(scenario.data)
			g.Expect(err).ToNot(HaveOccurred())
			var obj unstructured.Unstructured
			err = yaml.Unmarshal(yamlBytes, &obj)
			g.Expect(err).ToNot(HaveOccurred())

			healthStatus, err := hc.Check(obj)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(healthStatus.Status).To(Equal(scenario.healthStatus))
		})
	}
}
