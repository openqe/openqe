package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
)

func ExpandPath(path string) (string, error) {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func FileExists(filename string) bool {
	path, _ := ExpandPath(filename)
	_, err := os.Stat(path)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // file does not exist
	}
	return false // some other error (e.g. permission denied)
}

func AllTrue(str string) bool {
	words := strings.Fields(str) // splits on any whitespace
	for _, w := range words {
		if w != "True" {
			return false
		}
	}
	return true
}

func AllFalse(str string) bool {
	words := strings.Fields(str) // splits on any whitespace
	for _, w := range words {
		if w != "False" {
			return false
		}
	}
	return true
}

func ErrStack(out io.Writer, msg interface{}) {
	fmt.Fprintln(out, string(debug.Stack()))
	m := fmt.Sprintf("%v", msg)
	fmt.Fprintln(out, m)
}

// ExitError struct
type ExitError struct {
	Cmd    string
	StdErr string
	*exec.ExitError
}
