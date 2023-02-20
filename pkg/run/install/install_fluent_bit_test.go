package install

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
)

func TestMakeConfigOutputs(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name     string
		bucket   string
		port     int32
		expected string
		err      error
	}{
		{
			name:     "valid input",
			bucket:   "test-bucket",
			port:     8080,
			expected: "[OUTPUT]\n    Name s3\n    Match kube.*\n    bucket test-bucket\n    endpoint http://run-dev-bucket.gitops-run.svc:8080\n    tls Off\n    tls.verify Off\n    use_put_object true\n    preserve_data_ordering true\n    static_file_path true\n    total_file_size 1M\n    upload_timeout 15s\n    s3_key_format /fluent-bit-logs/$TAG[4].%Y%m%d%H%M%S",
			err:      nil,
		},
		{
			name:     "invalid port - too low",
			bucket:   "test-bucket",
			port:     0,
			expected: "",
			err:      fmt.Errorf("port 0 not between 1 and 65535"),
		},
		{
			name:     "invalid port - too high",
			bucket:   "test-bucket",
			port:     65536,
			expected: "",
			err:      fmt.Errorf("port 65536 not between 1 and 65535"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := makeConfigOutputs(test.bucket, test.port)
			if test.err != nil {
				g.Expect(err).To(MatchError(test.err))
			} else {
				g.Expect(err).To(BeNil())
			}
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestMakeFluentBitHelmRelease(t *testing.T) {
	g := NewGomegaWithT(t)

	name := "test-release"
	namespace := "test-ns"
	targetNamespace := "test-target-ns"
	bucketName := "test-bucket"
	bucketServerPort := int32(9000)

	configOutputs, err := makeConfigOutputs(bucketName, bucketServerPort)
	g.Expect(err).To(BeNil())

	expectedValues, _ := json.Marshal(map[string]interface{}{
		"env": []map[string]interface{}{
			{
				"name": "AWS_ACCESS_KEY_ID",
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": constants.RunDevBucketCredentials,
						"key":  "accesskey",
					},
				},
			},
			{
				"name": "AWS_SECRET_ACCESS_KEY",
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"name": constants.RunDevBucketCredentials,
						"key":  "secretkey",
					},
				},
			},
		},
		"config": map[string]interface{}{
			"inputs":  strings.TrimSpace(configInputs),
			"filters": strings.TrimSpace(configFilters),
			"outputs": configOutputs,
		},
	})

	release, err := makeFluentBitHelmRelease(name, namespace, targetNamespace, bucketName, bucketServerPort)
	g.Expect(err).To(BeNil())
	g.Expect(release.Name).To(Equal(name))
	g.Expect(release.Namespace).To(Equal(namespace))
	g.Expect(release.Spec.Chart.Spec.Chart).To(Equal("fluent-bit"))
	g.Expect(release.Spec.Chart.Spec.Version).To(Equal("*"))
	g.Expect(release.Spec.Chart.Spec.SourceRef.Kind).To(Equal("HelmRepository"))
	g.Expect(release.Spec.Chart.Spec.SourceRef.Name).To(Equal("fluent"))
	g.Expect(release.Spec.Chart.Spec.SourceRef.Namespace).To(Equal("test-ns"))
	g.Expect(release.Spec.Values.Raw).To(MatchJSON(expectedValues))
}
