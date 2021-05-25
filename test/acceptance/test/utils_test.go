package acceptance

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	require.True(t, FileExists("utils_test.go"))
	require.False(t, FileExists("imaginaryfile.txt"))
}
