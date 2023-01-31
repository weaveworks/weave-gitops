package install

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestMakeVClusterHelmReleaseAnnotations(t *testing.T) {
	g := NewGomegaWithT(t)

	hl, err := makeVClusterHelmRelease(
		"name",
		"namespace",
		"flux-system",
		"command", []string{"9999", "1111"},
		"automationKind")
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(hl.Name).To(Equal("name"))
	g.Expect(hl.Namespace).To(Equal("namespace"))

	values := map[string]interface{}{}
	g.Expect(json.Unmarshal(hl.Spec.Values.Raw, &values)).ToNot(HaveOccurred())

	annotations := values["annotations"].(map[string]interface{})
	g.Expect(annotations["run.weave.works/automation-kind"]).To(Equal("automationKind"))
	g.Expect(annotations["run.weave.works/namespace"]).To(Equal("namespace"))
	g.Expect(annotations["run.weave.works/command"]).To(Equal("command"))
	g.Expect(annotations["run.weave.works/port-forward"]).To(Equal("9999,1111"))
}
