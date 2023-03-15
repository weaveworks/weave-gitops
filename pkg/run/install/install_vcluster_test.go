package install

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestMakeVClusterHelmReleaseAnnotations(t *testing.T) {

	tests := []struct {
		name                  string
		portForwards          []string
		expectedAnnotations   map[string]string
		unexpectedAnnotations []string
	}{
		{
			name:         "port-forwards are properly annotated",
			portForwards: []string{"9999", "1111"},
			expectedAnnotations: map[string]string{
				"run.weave.works/port-forward":    "9999,1111",
				"run.weave.works/automation-kind": "automationKind",
				"run.weave.works/namespace":       "namespace",
				"run.weave.works/command":         "command",
			},
		},
		{
			name:                  "port-forwards are left out",
			unexpectedAnnotations: []string{"run.weave.works/port-forward"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			hl := makeVClusterHelmRelease(
				"name",
				"namespace",
				"flux-system",
				"command", tt.portForwards,
				"automationKind")

			g.Expect(hl.Name).To(Equal("name"))
			g.Expect(hl.Namespace).To(Equal("namespace"))

			values := map[string]interface{}{}
			g.Expect(json.Unmarshal(hl.Spec.Values.Raw, &values)).ToNot(HaveOccurred())

			annotations := values["annotations"].(map[string]interface{})

			for k, v := range tt.expectedAnnotations {
				g.Expect(annotations).To(HaveKeyWithValue(k, v))
			}

			for _, k := range tt.unexpectedAnnotations {
				g.Expect(annotations).NotTo(HaveKey(k))
			}
		})
	}

}
