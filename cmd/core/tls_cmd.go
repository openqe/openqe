package core

import (
	"github.com/gaol/openqe/pkg/tls"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func NewTLSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tls",
		Short:        "TLS oriented test utilities",
		SilenceUsage: true,
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	cmd.AddCommand(NewCAGenCommand())
	cmd.AddCommand(NewTLSGenCommand())
	return cmd
}

// ============    CA-GEN COMMAND     ==============================
func BindCAOptions(opts *tls.CAOptions, flags *flag.FlagSet) {
	flags.StringVar(&opts.Subject, "ca-subject", opts.Subject, "The CA certificate subject used to generate the TLS CA.")
	flags.StringVar(&opts.DNSName, "ca-dns-name", opts.DNSName, "The SAN used to generate the TLS CA.")
	flags.StringVar(&opts.CaKeyFile, "ca-key-file", opts.CaKeyFile, "The CA private key file path to be generated to.")
	flags.StringVar(&opts.CaCertFile, "ca-cert-file", opts.CaCertFile, "The CA certificate file path to be generated to.")
}

func NewCAGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ca-gen",
		Short:         "Generate CA key/cert pair to files",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	opts := tls.DefaultCAOptions()
	BindCAOptions(opts, cmd.Flags())
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := tls.GenerateCAToFiles(opts); err != nil {
			cmd.Printf("Failed to generate the CA key/cert pair: %s\n", err)
			return
		}
		cmd.Printf("CA generated to caKeyFile: %s, caCertFile: %s\n", opts.CaKeyFile, opts.CaCertFile)
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

func NewTLSGenCommand() *cobra.Command {
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
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := tls.GenerateTLSKeyCertPairToFiles(opts); err != nil {
			cmd.Printf("Failed to generate the TLS key/cert pair: %s\n", err)
			return
		}
		cmd.Printf("TLS key/cert paris gets generated to keyFile: %s, certFile: %s\n", opts.KeyFile, opts.CertFile)
	}
	return cmd
}
