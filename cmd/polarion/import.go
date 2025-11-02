package polarion

import (
	"fmt"

	"github.com/openqe/openqe/pkg/polarion"
	"github.com/spf13/cobra"
)

type ImportOptions struct {
	ConfigFile     string
	TestCasesFile  string
	DryRun         bool
	Verbose        bool
	TestConnection bool
}

func NewImportCommand() *cobra.Command {
	opts := &ImportOptions{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import test cases to Polarion",
		Long: `Import test cases to Polarion from a YAML file.

Examples:
  # Import using default config (config.local.yaml)
  openqe polarion import

  # Import using specific config file
  openqe polarion import --config my_config.yaml

  # Import using specific test cases file
  openqe polarion import --test-cases test_cases.yaml

  # Dry run to see what would be created
  openqe polarion import --dry-run

  # Test connection only
  openqe polarion import --test-connection
`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create importer instance
			importer, err := polarion.NewImporter(opts.ConfigFile, opts.Verbose)
			if err != nil {
				return fmt.Errorf("failed to create importer: %w", err)
			}

			// Override test cases file if provided
			if opts.TestCasesFile != "" {
				importer.SetTestCasesFile(opts.TestCasesFile)
			}

			// Test connection only
			if opts.TestConnection {
				return importer.TestConnection()
			}

			// Import all test cases
			return importer.ImportAll(opts.DryRun)
		},
	}

	cmd.Flags().StringVarP(&opts.ConfigFile, "config", "c", "config.local.yaml", "Path to configuration file")
	cmd.Flags().StringVarP(&opts.TestCasesFile, "test-cases", "t", "", "Path to test cases file (overrides config)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Perform a dry run without actually creating test cases")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose logging")
	cmd.Flags().BoolVar(&opts.TestConnection, "test-connection", false, "Only test the connection to Polarion server")

	return cmd
}
