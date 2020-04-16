package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func checkStringContains(t *testing.T, got, expected string) {
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}
}

func TestRootCommandUnknownCommand(t *testing.T) {
	output, err := executeCommand(rootCmd, "one", "two")
	checkStringContains(t, output, "Error: unknown command \"one\" for \"nancy\"")
	checkStringContains(t, err.Error(), "unknown command \"one\" for \"nancy\"")
}

const envExitTest = "BE_EXIT_TEST"

func TestRootCommandNoArgsNoStdIn(t *testing.T) {
	// run test in separate process
	if os.Getenv(envExitTest) == "true" {
		_, _ = executeCommand(rootCmd, "")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
	cmd.Env = append(os.Environ(), envExitTest+"=true")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		checkStringContains(t, err.Error(), "exit status 2")
		return
	}
	t.Fatalf("process ran with err %v, want exit status 2", err)
}

func TestRootCommandOssi(t *testing.T) {
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }() // Restore original Stdin

	// setup fake Stdin
	tmpfile, err := ioutil.TempFile("", "example")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte("Tom")); err != nil {
		assert.NoError(t, err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		assert.NoError(t, err)
	}
	os.Stdin = tmpfile
	defer tmpfile.Close()

	// run test in separate process
	if os.Getenv("BE_CRASHER") == "1" {
		_, _ = executeCommand(rootCmd, "")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestRootCommandOssi")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err = cmd.Run()
	assert.NoError(t, err)
}
