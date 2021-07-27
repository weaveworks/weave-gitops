package gitops

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostRunDefined(t *testing.T) {
	assert.NotNil(t, installCmd.PostRun, "PostRun should be defined for install")
	assert.NotNil(t, uinstallCmd.PostRun, "PostRun should be defined for uninstall")
}
