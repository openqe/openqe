package openshift

import "github.com/openqe/openqe/pkg/exec"

// This file contains functions that will use oc client CLI instead of go client
func OC_Get(kubeconfig string, verbose bool, args ...string) (string, error) {
	finalArgs := []string{"--kubeconfig", kubeconfig, "get"}
	finalArgs = append(finalArgs, args...)
	cli := &exec.CLI{
		ExecPath: "oc",
		Args:     finalArgs,
		Verbose:  verbose,
	}
	output, err := cli.Execute()
	if err != nil {
		return "", err
	}
	return output, nil
}
