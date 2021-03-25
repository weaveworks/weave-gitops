package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	result := test()
	require.Equal(t, "test:677289984", result)
}
