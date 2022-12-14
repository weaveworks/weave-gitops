package session

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mockList struct {
	client.Client
}

var _ client.Client = &mockList{}

func (m *mockList) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error {
	switch list := list.(type) {
	case *v1.StatefulSetList:
		*list = v1.StatefulSetList{
			Items: []v1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "name",
						Namespace: "namespace",
						Annotations: map[string]string{
							"run.weave.works/command":      "command",
							"run.weave.works/cli-version":  "cli-version",
							"run.weave.works/port-forward": "9999,1111",
							"run.weave.works/namespace":    "flux-system",
						},
					},
				},
			},
		}
	}

	return nil
}

func TestList(t *testing.T) {
	g := NewGomegaWithT(t)
	list, err := List(&mockList{}, "namespace")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(list).ToNot(BeNil())
	g.Expect(list).To(HaveLen(1))
	g.Expect(list[0].SessionName).To(Equal("name"))
	g.Expect(list[0].SessionNamespace).To(Equal("namespace"))
	g.Expect(list[0].PortForward).To(Equal([]string{"9999", "1111"}))
	g.Expect(list[0].Command).To(Equal("command"))
	g.Expect(list[0].CliVersion).To(Equal("cli-version"))
	g.Expect(list[0].Namespace).To(Equal("flux-system"))
}
