/*
Copyright © 2025 Lin Gao <aoingl@gmail.com>
*/
package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/openqe/openqe/cmd/auth"
	core "github.com/openqe/openqe/cmd/core"
	"github.com/openqe/openqe/cmd/openshift"
	"github.com/openqe/openqe/cmd/polarion"
	"github.com/spf13/cobra"
)

const NAME string = "openqe"

type ProjectInformation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

//go:embed project.yaml
var projectInfoBytes []byte

func getVersion() string {
	projectInfo := &ProjectInformation{}
	if err := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(projectInfoBytes), 100).Decode(&projectInfo); err != nil {
		panic(err)
	}
	return projectInfo.Version
}

// GetRevision returns the overall codebase version. It's for detecting
// what code a binary was built from.
func GetRevision() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "<unknown>"
	}
	for _, setting := range bi.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return "<unknown>"
}

func VersionString() string {
	return fmt.Sprintf("%s version: %s, (Revision: %s)", NAME, getVersion(), GetRevision())
}

func VersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "version",
		Short:        "Show current version",
		SilenceUsage: true,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Printf("%s\n", VersionString())
	}
	return cmd

}
func addCommands(rootCommand *cobra.Command) {
	rootCommand.AddCommand(VersionCommand())
	rootCommand.AddCommand(core.NewTLSCommand())
	rootCommand.AddCommand(openshift.NewCommand())
	rootCommand.AddCommand(auth.NewAuthCommand())
	rootCommand.AddCommand(polarion.NewCommand())
	rootCommand.AddCommand(core.NewDocCommand(rootCommand))
}

func main() {
	cmd := &cobra.Command{
		Use:               NAME,
		SilenceUsage:      true,
		TraverseChildren:  true,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	cmd.Version = VersionString()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	addCommands(cmd)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		<-sigs
		fmt.Fprintln(os.Stderr, "\nAborted...")
		cancel()
	}()

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
