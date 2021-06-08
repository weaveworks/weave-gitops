package utils

// Utilities to run external commands.

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/shims"
)

type CallOperation int

const (
	CallCommandOp CallOperation = iota
	CallCommandSilentlyOp
	CallCommandSeparatingOutputStreamsOp
	CallCommandForEffectOp
	CallCommandForEffectWithDebugOp
	CallCommandForEffectWithInputPipeOp
	CallCommandForEffectWithInputPipeAndDebugOp
)

//type Behavior func(args ...interface{}) ([]byte, []byte, error)

var (
	behaviors = make([]interface{}, int(CallCommandForEffectWithInputPipeAndDebugOp+1))
)

func processMocks(op CallOperation, cmdstr string) (bool, []byte, []byte, error) {
	if behavior := behaviors[op]; behavior != nil {
		if stdout, stderr, err := behavior.(func(...interface{}) ([]byte, []byte, error))(cmdstr); err != nil {
			return true, nil, nil, err
		} else {
			return true, stdout, stderr, nil
		}
	}
	return false, nil, nil, nil
}

// CallCommand will run an external command, displaying its output interactively and return its output.
func CallCommand(cmdstr string) ([]byte, error) {
	if processed, outval, _, err := processMocks(CallCommandOp, cmdstr); processed {
		return outval, err
	}

	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	var out strings.Builder
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	stdoutScanner := bufio.NewScanner(stdoutReader)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for stdoutScanner.Scan() {
			data := stdoutScanner.Text()
			fmt.Println(data)
			out.WriteString(data)
			out.WriteRune('\n')
		}
	}()

	stderrScanner := bufio.NewScanner(stderrReader)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for stderrScanner.Scan() {
			data := stderrScanner.Text()
			fmt.Println(data)
			out.WriteString(data)
			out.WriteRune('\n')
		}
	}()

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	wg.Wait()
	err = cmd.Wait()

	return []byte(out.String()), err
}

func CallCommandSilently(cmdstr string) ([]byte, error) {
	if processed, val, _, err := processMocks(CallCommandSilentlyOp, cmdstr); processed {
		return val, err
	}

	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	return cmd.CombinedOutput()
}

func CallCommandSeparatingOutputStreams(cmdstr string) ([]byte, []byte, error) {
	if processed, outval, errval, err := processMocks(CallCommandSeparatingOutputStreamsOp, cmdstr); processed {
		return outval, errval, err
	}

	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func CallCommandForEffect(cmdstr string) error {
	if processed, _, _, err := processMocks(CallCommandForEffectOp, cmdstr); processed {
		return err
	}

	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	return cmd.Run()
}

func CallCommandForEffectWithDebug(cmdstr string) error {
	if processed, _, _, err := processMocks(CallCommandForEffectWithDebugOp, cmdstr); processed {
		return err
	}

	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CallCommandForEffectWithInputPipe(cmdstr, input string) error {
	if processed, _, _, err := processMocks(CallCommandForEffectWithInputPipeOp, cmdstr); processed {
		return err
	}

	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	inpipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		_, _ = io.WriteString(inpipe, input)
		inpipe.Close()
	}()
	return cmd.Run()
}

func CallCommandForEffectWithInputPipeAndDebug(cmdstr, input string) error {
	if processed, _, _, err := processMocks(CallCommandForEffectWithInputPipeAndDebugOp, cmdstr); processed {
		return err
	}

	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	inpipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		_, _ = io.WriteString(inpipe, input)
		inpipe.Close()
	}()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Escape(cmd string) string {
	return strings.ReplaceAll(cmd, "'", "'\"'\"'")
}

func OverrideBehavior(callOp CallOperation, behavior func(args ...interface{}) ([]byte, []byte, error)) override.Override {
	location := &behaviors[callOp]
	return override.Override{Handler: location, Mock: behavior, Original: behaviors[callOp]}
}

func OverrideFailure(callOp CallOperation) override.Override {
	location := &behaviors[callOp]
	return override.Override{Handler: location, Mock: func(args ...interface{}) ([]byte, []byte, error) {
		fmt.Println("failing for ", args)
		shims.Exit(1)
		return nil, nil, nil
	},
		Original: behaviors[callOp],
	}
}

func OverrideIgnore(callOp CallOperation) override.Override {
	location := &behaviors[callOp]
	return override.Override{Handler: location, Mock: func(args ...interface{}) ([]byte, []byte, error) {
		fmt.Println("ignoring ", args)
		return nil, nil, nil
	},
		Original: behaviors[callOp],
	}
}

func WithBehaviorFor(callOp CallOperation, behavior func(args ...interface{}) ([]byte, []byte, error), action func() ([]byte, []byte, error)) ([]byte, []byte, error) {
	existingBehavior := behaviors[callOp]
	behaviors[callOp] = behavior
	defer func() {
		behaviors[callOp] = existingBehavior
	}()
	return action()
}

func WithResultsFrom(callOp CallOperation, outvalue []byte, errvalue []byte, err error, action func() ([]byte, []byte, error)) ([]byte, []byte, error) {
	return WithBehaviorFor(
		callOp,
		func(args ...interface{}) ([]byte, []byte, error) {
			return outvalue, errvalue, err
		},
		action)
}
