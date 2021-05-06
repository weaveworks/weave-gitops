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
)

// CallCommand will run an external command, displaying its output interactively and return its output.
func CallCommand(cmdstr string) ([]byte, error) {
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
	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	return cmd.CombinedOutput()
}

func CallCommandSeparatingOutputStreams(cmdstr string) ([]byte, []byte, error) {
	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func CallCommandForEffect(cmdstr string) error {
	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	return cmd.Run()
}

func CallCommandForEffectWithDebug(cmdstr string) error {
	cmd := exec.Command("sh", "-c", Escape(cmdstr))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func CallCommandForEffectWithInputPipe(cmdstr, input string) error {
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
