package acceptance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	require.True(t, FileExists("utils_test.go"))
	require.False(t, FileExists("imaginaryfile.txt"))
}
