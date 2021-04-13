package add

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gotest.tools/assert"
)

func TestAddHelm(t *testing.T) {
	expectedManifests := `---
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
	name: podinfo-wego
	namespace: flux-system
spec:
	interval: 30s
	ref:
		branch: master
	url: https://github.com/stefanprodan/podinfo

---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
	name: podinfo-wego
	namespace: flux-system
spec:
	chart:
		spec:
			chart: ./charts/podinfo
			sourceRef:
				kind: GitRepository
				name: podinfo-wego
	interval: 5m0s

`

	params = paramSet{
		url:           "https://github.com/stefanprodan/podinfo",
		name:          "podinfo-wego",
		branch:        "master",
		path:          "./charts/podinfo",
		manifestsKind: "helm",
	}

	output := captureStdout(func() {
		runCmd(&cobra.Command{}, []string{})
	})

	assert.Equal(t, strings.ReplaceAll(expectedManifests, "\t", "  "), output)
}

func captureStdout(f func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()

	os.Stdout = w
	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	os.Stdout = orig

	return buf.String()
}
