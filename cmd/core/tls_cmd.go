package core

import (
	"fmt"

	"github.com/openqe/openqe/pkg/common"
	"github.com/openqe/openqe/pkg/tls"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func NewTLSCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tls",
		Short:        "TLS oriented test utilities",
		SilenceUsage: true,
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	cmd.AddCommand(NewCAGenCommand(globalOpts))
	cmd.AddCommand(NewTLSGenCommand(globalOpts))
	cmd.AddCommand(NewCACheckCommand(globalOpts))
	return cmd
}

// ============    CA-GEN COMMAND     ==============================
func BindCAOptions(opts *tls.CAOptions, flags *flag.FlagSet) {
	flags.StringVar(&opts.Subject, "ca-subject", opts.Subject, "The CA certificate subject used to generate the TLS CA.")
	flags.StringVar(&opts.DNSName, "ca-dns-name", opts.DNSName, "The SAN used to generate the TLS CA.")
	flags.StringVar(&opts.CaKeyFile, "ca-key-file", opts.CaKeyFile, "The CA private key file path to be generated to.")
	flags.StringVar(&opts.CaCertFile, "ca-cert-file", opts.CaCertFile, "The CA certificate file path to be generated to.")
}

func NewCAGenCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ca-gen",
		Short:         "Generate CA key/cert pair to files",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	opts := tls.DefaultCAOptions()
	BindCAOptions(opts, cmd.Flags())
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logger := common.NewLoggerFromOptions(globalOpts, "TLS")

		if err := tls.GenerateCAToFiles(opts); err != nil {
			return fmt.Errorf("Failed to generate the CA key/cert pair: %v", err)
		}
		logger.Info("CA generated to caKeyFile: %s, caCertFile: %s", opts.CaKeyFile, opts.CaCertFile)
		return nil
	}
	return cmd
}

// ============    TLS CERT-GEN COMMAND     ==============================

func BindPKIOptions(opts *tls.PKIOptions, flags *flag.FlagSet) {
	BindCAOptions(opts.CaGenOpt, flags)
	flags.StringVar(&opts.CertFile, "tls-cert-file", opts.CertFile, "The file path of the TLS certificate to be generated to.")
	flags.StringVar(&opts.KeyFile, "tls-key-file", opts.KeyFile, "The file path of the TLS private key to be generated to.")
	flags.StringVar(&opts.Subject, "subject", opts.Subject, "The TLS certificate subject.")
	flags.StringVar(&opts.DNSName, "dns-name", opts.DNSName, "The SAN added to the TLS certificate.")
}

func NewTLSGenCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert-gen",
		Short: "Generate TLS key/cert pair to files, signed by a given CA",
		Long: `Generate TLS key/cert pair to files, signed by a given CA.
The CA key/cert files must be provided to sign the generated TLS certificate.
You can use the 'tls ca-gen' command to generate a CA key/cert pair for testing purpose.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	opts := tls.DefaultPKIOptions()
	BindPKIOptions(opts, cmd.Flags())
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logger := common.NewLoggerFromOptions(globalOpts, "TLS")

		if err := tls.GenerateTLSKeyCertPairToFiles(opts); err != nil {
			return fmt.Errorf("Failed to generate the TLS key/cert pair: %s", err)
		}
		logger.Info("TLS key/cert pairs generated to keyFile: %s, certFile: %s", opts.KeyFile, opts.CertFile)
		return nil
	}
	return cmd
}

// ============    CA-CHECK COMMAND     ==============================

type CACheckOptions struct {
	CACertFile   string
	CABundleFile string
}

func NewCACheckCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ca-check",
		Short: "Check if a CA certificate is included in a CA bundle file",
		Long: `Check if a CA certificate file is included in a CA bundle file.
This command will return success (exit code 0) if the CA certificate is found in the bundle,
or failure (exit code 1) if it is not found.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	opts := &CACheckOptions{
		CABundleFile: "/etc/pki/tls/certs/ca-bundle.crt", // default ca bundle file
	}
	cmd.Flags().StringVar(&opts.CACertFile, "ca-cert-file", opts.CACertFile, "The CA certificate file to check")
	cmd.Flags().StringVar(&opts.CABundleFile, "ca-bundle-file", opts.CABundleFile, "The CA bundle file to check against")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logger := common.NewLoggerFromOptions(globalOpts, "TLS")

		if opts.CACertFile == "" {
			cmd.Usage()
			return fmt.Errorf("Error: --ca-cert-file is required")
		}
		if opts.CABundleFile == "" {
			cmd.Usage()
			return fmt.Errorf("Error: --ca-bundle-file is required")
		}

		found, err := tls.CheckCACertInBundle(opts.CACertFile, opts.CABundleFile)
		if err != nil {
			return fmt.Errorf("Error checking CA certificate in bundle: %v", err)
		}

		if found {
			logger.Info("CA certificate found in bundle")
		} else {
			return fmt.Errorf("CA certificate NOT found in bundle")
		}
		return nil
	}
	return cmd
}
