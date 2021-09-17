package version

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestPostRunDefined(t *testing.T) {
	assert.NotNil(t, Cmd.PostRun, "PostRun should be defined")
}

func TestCheckpointMissingParent(t *testing.T) {
	c := &cobra.Command{
		Use: "test",
	}
	assert.Empty(t, CheckpointParamsWithFlags(CheckpointParams(), c).Flags,
		"should have empty flags when command doesn't have a parent")
}

func TestCheckpointParamsWithFlagsNoParms(t *testing.T) {
	c := &cobra.Command{
		Use: "test",
	}
	assert.Empty(t, CheckpointParamsWithFlags(nil, c).Flags,
		"should have empty flags when no params passed")
}

func TestCheckpointParamsWithFlagsUseParsing(t *testing.T) {
	p1 := &cobra.Command{
		Use: "obj asdf asdf;kljasd",
	}
	p2 := &cobra.Command{
		Use: "cmd asdf asdf;kljasd",
	}
	p1.AddCommand(p2)

	expectedRes := map[string]string{
		"object":  "obj",
		"command": "cmd",
	}
	assert.Equal(t, expectedRes, CheckpointParamsWithFlags(nil, p2).Flags,
		"should only have the first word of the use field")
}

func TestCheckpointParamsWithFlagsAndWegoParent(t *testing.T) {
	p1 := &cobra.Command{
		Use: "wego",
	}
	p2 := &cobra.Command{
		Use: "p1.p1 asdf asdf;kljasd",
	}
	p1.AddCommand(p2)
	assert.Empty(t, CheckpointParamsWithFlags(nil, p2).Flags,
		"should only have the first word of the use field")
}
