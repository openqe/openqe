package exec

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type CLI struct {
	execPath string
	args     []string
	stdin    *bytes.Buffer
	stdout   io.Writer
	stderr   io.Writer
	verbose  bool
}

// Execute the command and returns stdout/stderr combined into one string
func (c *CLI) Execute() (string, error) {
	if c.verbose {
		fmt.Printf("DEBUG: %s\n", c.String())
	}
	cmd := exec.Command(c.execPath, c.args...)
	cmd.Stdin = c.stdin
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if err != nil {
		return "", err
	}
	return trimmed, nil
}

// Execute the command and returns the stdout/stderr output as separate strings
func (c *CLI) Execute2() (string, string, error) {
	if c.verbose {
		fmt.Printf("DEBUG: %s\n", c.String())
	}
	cmd := exec.Command(c.execPath, c.args...)
	cmd.Stdin = c.stdin
	var stdErrBuff, stdOutBuff bytes.Buffer
	cmd.Stdout = &stdOutBuff
	cmd.Stderr = &stdErrBuff
	err := cmd.Run()

	stdOutBytes := stdOutBuff.Bytes()
	stdErrBytes := stdErrBuff.Bytes()
	stdOut := strings.TrimSpace(string(stdOutBytes))
	stdErr := strings.TrimSpace(string(stdErrBytes))
	if err != nil {
		return "", "", err
	}
	return stdOut, stdErr, nil
}

func (c *CLI) String() string {
	return fmt.Sprintf("ExecutePath: %s, Args: %s", c.execPath, strings.Join(c.args, " "))
}
