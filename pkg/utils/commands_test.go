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

func TestWithBehaviorFor(t *testing.T) {
	success := false
	val := []byte("This is a result string")
	out, _, _ :=
		utils.WithBehaviorFor(utils.CallCommandOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				out, err := utils.CallCommand("echo cornhusk")
				assert.NoError(t, err)
				return out, nil, err
			})
	assert.Equal(t, val, out)
	assert.True(t, success)

	success = false
	out, _, _ =
		utils.WithBehaviorFor(utils.CallCommandSilentlyOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				out, err := utils.CallCommandSilently("echo cornhusk")
				assert.NoError(t, err)
				return out, nil, err
			})
	assert.Equal(t, val, out)
	assert.True(t, success)

	success = false
	out, _, _ =
		utils.WithBehaviorFor(utils.CallCommandSeparatingOutputStreamsOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				sout, serr, err := utils.CallCommandSeparatingOutputStreams("echo cornhusk")
				assert.NoError(t, err)
				return sout, serr, err
			})
	assert.Equal(t, val, out)
	assert.True(t, success)

	success = false
	_, _, err :=
		utils.WithBehaviorFor(utils.CallCommandForEffectOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				err := utils.CallCommandForEffect("echo cornhusk")
				assert.NoError(t, err)
				return nil, nil, nil
			})
	assert.NoError(t, err)
	assert.True(t, success)

	success = false
	_, _, err =
		utils.WithBehaviorFor(utils.CallCommandForEffectWithDebugOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				err := utils.CallCommandForEffectWithDebug("echo cornhusk")
				assert.NoError(t, err)
				return nil, nil, nil
			})
	assert.NoError(t, err)
	assert.True(t, success)

	success = false
	_, _, err =
		utils.WithBehaviorFor(utils.CallCommandForEffectWithInputPipeOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				err := utils.CallCommandForEffectWithInputPipe("echo cornhusk", "x")
				assert.NoError(t, err)
				return nil, nil, nil
			})
	assert.NoError(t, err)
	assert.True(t, success)

	success = false
	_, _, err =
		utils.WithBehaviorFor(utils.CallCommandForEffectWithInputPipeAndDebugOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				err := utils.CallCommandForEffectWithInputPipeAndDebug("echo cornhusk", "x")
				assert.NoError(t, err)
				return nil, nil, nil
			})
	assert.NoError(t, err)
	assert.True(t, success)

	// Test case where another behavior was already defined
	success = false
	out, _, _ =
		utils.WithBehaviorFor(utils.CallCommandSilentlyOp,
			func(args ...interface{}) ([]byte, []byte, error) {
				success = true
				return val, nil, nil
			},
			func() ([]byte, []byte, error) {
				return utils.WithBehaviorFor(utils.CallCommandSilentlyOp,
					func(args ...interface{}) ([]byte, []byte, error) {
						success = true
						return val, nil, nil
					},
					func() ([]byte, []byte, error) {
						out, err := utils.CallCommandSilently("echo cornhusk")
						assert.NoError(t, err)
						return out, nil, err
					})
			})
	assert.Equal(t, val, out)
	assert.True(t, success)
}
