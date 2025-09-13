package core

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func NewDocCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "doc",
		Short:        "Documentation related commands",
		SilenceUsage: true,
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	cmd.AddCommand(NewCobraDocGenCmd(rootCmd))
	return cmd
}

type DocGenOptions struct {
	Output string
}

var opts = &DocGenOptions{}

func NewCobraDocGenCmd(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cobra-doc-gen",
		Short:        "Generate the markdown documentation for the CLI",
		SilenceUsage: true,
	}
	cmd.Flags().StringVar(&opts.Output, "output", opts.Output, "The CA certificate subject used to generate the TLS CA.")
	cmd.MarkFlagRequired("output")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := doc.GenMarkdownTree(rootCmd, opts.Output)
		if err != nil {
			return fmt.Errorf("Failed to generate the documentation: %v\n", err)
		}
		cmd.Printf("The documentation is generated to %s\n", opts.Output)
		return nil
	}
	return cmd
}
