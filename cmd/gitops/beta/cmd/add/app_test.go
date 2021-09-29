package add

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateToNewDirStructure(t *testing.T) {
	tests := []struct {
		orig string
		exp  string
	}{
		{"foo", "foo"},
		{"apps/foo/foo.yaml", ".weave-gitops/apps/foo/foo.yaml"},
		{".wego/apps/foo/foo.yaml", ".weave-gitops/apps/foo/foo.yaml"},
		{"targets/mycluster/foo/deploy.yaml", ".weave-gitops/apps/foo/deploy.yaml"},
		{".wego/targets/mycluster/foo/source.yaml", ".weave-gitops/apps/foo/source.yaml"},
		{"", ""},
	}
	for _, i := range tests {
		assert.Equal(t, i.exp, migrateToNewDirStructure(i.orig))

	}
}
