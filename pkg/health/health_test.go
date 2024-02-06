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
		{
			data:         "testdata/daemonset-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/daemonset-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/statefulset-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/statefulset-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/job-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/job-unhealthy.yaml",
			healthStatus: HealthStatusUnhealthy,
		},
		{
			data:         "testdata/hpa-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/hpa-unhealthy.yaml",
			healthStatus: HealthStatusUnhealthy,
		},
		{
			data:         "testdata/hpa-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/ingress-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/ingress-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/pvc-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/pvc-unhealthy.yaml",
			healthStatus: HealthStatusUnhealthy,
		},
		{
			data:         "testdata/pvc-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/svc-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/svc-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/kstatus-healthy.yaml",
			healthStatus: HealthStatusHealthy,
		},
		{
			data:         "testdata/kstatus-progressing.yaml",
			healthStatus: HealthStatusProgressing,
		},
		{
			data:         "testdata/kstatus-unhealty.yaml",
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
