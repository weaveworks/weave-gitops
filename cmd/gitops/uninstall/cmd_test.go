package uninstall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostRunDefined(t *testing.T) {
	assert.NotNil(t, Cmd.PostRun, "PostRun should be defined for uninstall")
}
