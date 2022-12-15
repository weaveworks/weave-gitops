package session

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mockGet struct {
	client.Client
}

var _ client.Client = &mockGet{}

func (m *mockGet) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	switch obj := obj.(type) {
	case *v1.StatefulSet:
		*obj = v1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
				Annotations: map[string]string{
					"run.weave.works/command":      "command",
					"run.weave.works/cli-version":  "cli-version",
					"run.weave.works/port-forward": "9999,1111",
					"run.weave.works/namespace":    "flux-system",
				},
			},
		}
	}

	return nil
}

func TestGet(t *testing.T) {
	g := NewGomegaWithT(t)
	is, err := Get(&mockGet{}, "name", "namespace")
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(is).ToNot(BeNil())
	g.Expect(is.SessionName).To(Equal("name"))
	g.Expect(is.SessionNamespace).To(Equal("namespace"))
	g.Expect(is.PortForward).To(Equal([]string{"9999", "1111"}))
	g.Expect(is.Command).To(Equal("command"))
	g.Expect(is.CliVersion).To(Equal("cli-version"))
	g.Expect(is.Namespace).To(Equal("flux-system"))
}
