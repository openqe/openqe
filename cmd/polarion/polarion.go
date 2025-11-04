package polarion

import (
	"github.com/openqe/openqe/pkg/common"
	"github.com/spf13/cobra"
)

func NewCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "polarion",
		Short:         "Polarion test case management utilities",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(NewImportCommand(globalOpts))
	cmd.AddCommand(NewInspectCommand(globalOpts))
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	return cmd
}
