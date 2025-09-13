package exec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/openqe/openqe/pkg/utils"
)

type CLI struct {
	ExecPath string
	Args     []string
	Stdin    *bytes.Buffer
	Stdout   io.Writer
	Stderr   io.Writer
	Verbose  bool
}

// Execute the command and returns stdout/stderr combined into one string
func (c *CLI) Execute() (string, error) {
	if c.Verbose {
		fmt.Printf("DEBUG: %s\n", c.String())
	}
	cmd := exec.Command(c.ExecPath, c.Args...)
	cmd.Stdin = c.Stdin
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	switch err.(type) {
	case nil:
		c.Stdout = bytes.NewBuffer(out)
		return trimmed, nil
	case *exec.ExitError:
		return trimmed, &utils.ExitError{ExitError: err.(*exec.ExitError), Cmd: c.ExecPath + " " + strings.Join(c.Args, " "), StdErr: trimmed}
	default:
		utils.ErrStack(os.Stderr, fmt.Errorf("unable to execute %q: %v", c.ExecPath, err))
		// unreachable code
		return "", nil
	}
}

func (c *CLI) String() string {
	return fmt.Sprintf("ExecutePath: %s, Args: %s", c.ExecPath, strings.Join(c.Args, " "))
}
