package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

func TestCallCommand(t *testing.T) {
	assert := assert.New(t)

	output, err := utils.CallCommand(`echo "stdout" && >&2 echo "stderr"`)
	assert.NoError(err)
	assert.Contains(string(output), "stderr")
	assert.Contains(string(output), "stdout")

	output, err = utils.CallCommand(`exit 1`)
	assert.EqualError(err, "exit status 1")
	assert.Equal("", string(output))
}

func TestCallCommandSilently(t *testing.T) {
	assert := assert.New(t)

	output, err := utils.CallCommandSilently(`echo "stdout" && >&2 echo "stderr"`)
	assert.NoError(err)
	assert.Contains(string(output), "stderr")
	assert.Contains(string(output), "stdout")

	output, err = utils.CallCommandSilently(`exit 1`)
	assert.EqualError(err, "exit status 1")
	assert.Equal("", string(output))
}

func TestCallCommandSeparatingOutputStreams(t *testing.T) {
	assert := assert.New(t)

	stdout, stderr, err := utils.CallCommandSeparatingOutputStreams(`echo "stdout" && >&2 echo "stderr"`)
	assert.NoError(err)
	assert.Equal("stdout\n", string(stdout))
	assert.Equal("stderr\n", string(stderr))

	_, _, err = utils.CallCommandSeparatingOutputStreams(`exit 1`)
	assert.EqualError(err, "exit status 1")
}

func TestCallCommandForEffect(t *testing.T) {
	assert := assert.New(t)

	err := utils.CallCommandForEffect(`echo "foo"`)
	assert.NoError(err)

	err = utils.CallCommandForEffect(`exit 1`)
	assert.EqualError(err, "exit status 1")
}

func TestCallCommandForEffectWithDebug(t *testing.T) {
	assert := assert.New(t)

	err := utils.CallCommandForEffectWithDebug(`echo "foo"`)
	assert.NoError(err)

	err = utils.CallCommandForEffectWithDebug(`exit 1`)
	assert.EqualError(err, "exit status 1")
}

func TestCallCommandForEffectWithInputPipe(t *testing.T) {
	assert := assert.New(t)

	err := utils.CallCommandForEffectWithInputPipe(`cat -`, "stuff\n")
	assert.NoError(err)

	err = utils.CallCommandForEffectWithInputPipe(`exit 1`, "stuff\n")
	assert.EqualError(err, "exit status 1")
}

func TestCallCommandForEffectWithInputPipeAndDebug(t *testing.T) {
	assert := assert.New(t)

	err := utils.CallCommandForEffectWithInputPipeAndDebug(`cat -`, "stuff\n")
	assert.NoError(err)

	err = utils.CallCommandForEffectWithInputPipeAndDebug(`exit 1`, "stuff\n")
	assert.EqualError(err, "exit status 1")
}

func TestEscape(t *testing.T) {
	assert := assert.New(t)

	str := utils.Escape(`'test'`)
	assert.Equal(`'"'"'test'"'"'`, str)
}
