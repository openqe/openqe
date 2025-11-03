package polarion

import (
	"fmt"

	"github.com/openqe/openqe/pkg/polarion"
	"github.com/spf13/cobra"
)

type InspectOptions struct {
	ConfigFile string
	WorkItemID string
	Verbose    bool
}

func NewInspectCommand() *cobra.Command {
	opts := &InspectOptions{}

	cmd := &cobra.Command{
		Use:   "inspect [work-item-id]",
		Short: "Inspect an existing work item to see its structure",
		Long: `Inspect an existing work item in Polarion to see what fields and values it contains.
This is useful for debugging field mapping issues.

Examples:
  # Inspect a work item
  openqe polarion inspect TEST-123

  # Inspect with custom config
  openqe polarion inspect TEST-123 --config my_config.yaml
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.WorkItemID = args[0]

			// Create importer instance to reuse config/client
			importer, err := polarion.NewImporter(opts.ConfigFile, opts.Verbose)
			if err != nil {
				return fmt.Errorf("failed to create importer: %w", err)
			}

			// Get the work item
			return importer.InspectWorkItem(opts.WorkItemID)
		},
	}

	cmd.Flags().StringVarP(&opts.ConfigFile, "config", "c", "config.local.yaml", "Path to configuration file")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose logging")

	return cmd
}
