package polarion

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "polarion",
		Short:         "Polarion test case management utilities",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(NewImportCommand())
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	return cmd
}
