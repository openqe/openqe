package core

import (
	"github.com/gaol/openqe/pkg/tls"
	"github.com/spf13/cobra"
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
type CAGenOptions struct {
	subject    string
	dnsName    string
	caKeyFile  string
	caCertFile string
}

func NewCAGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "ca-gen",
		Short:         "Generate CA key/cert pair to files",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	var opts CAGenOptions
	cmd.Flags().StringVar(&opts.subject, "subject", "C=China, O=OpenShift, OU=Hypershift QE, CN=default-ca", "The CA certificate subject used to generate the TLS CA.")
	cmd.Flags().StringVar(&opts.dnsName, "dns-name", "openqe.github.io", "The SAN used to generate the TLS CA.")
	cmd.Flags().StringVar(&opts.caKeyFile, "ca-key-file", "ca.key", "The CA private key file path to be generated to.")
	cmd.Flags().StringVar(&opts.caCertFile, "ca-cert-file", "ca.crt", "The CA certificate file path to be generated to.")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := tls.GenerateCAToFiles(opts.subject, opts.dnsName, opts.caKeyFile, opts.caCertFile); err != nil {
			cmd.Printf("Failed to generate the CA key/cert pair: %s\n", err)
			return
		}
		cmd.Printf("CA generated to caKeyFile: %s, caCertFile: %s\n", opts.caKeyFile, opts.caCertFile)
	}
	return cmd
}

// ============    TLS CERT-GEN COMMAND     ==============================
type PKIGenOptions struct {
	subject    string
	dnsName    string
	caKeyFile  string
	caCertFile string
	certFile   string
	keyFile    string
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

	var opts PKIGenOptions
	cmd.Flags().StringVar(&opts.subject, "subject", "C=China, O=OpenShift, OU=Hypershift QE, CN=default-server", "The TLS certificate subject.")
	cmd.Flags().StringVar(&opts.dnsName, "dns-name", "server.openqe.github.io", "The SAN used to generate the TLS certificate.")
	cmd.Flags().StringVar(&opts.caKeyFile, "ca-key-file", "ca.key", "The file from which to read the TLS CA private key.")
	cmd.Flags().StringVar(&opts.caCertFile, "ca-cert-file", "ca.crt", "The file from which to read TLS CA certificate file.")
	cmd.Flags().StringVar(&opts.keyFile, "tls-key-file", "tls.key", "The file path of the TLS private key to be generated to.")
	cmd.Flags().StringVar(&opts.certFile, "tls-cert-file", "tls.crt", "The file path of the TLS certificate to be generated to.")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := tls.GenerateTLSKeyCertPairToFiles(opts.subject, opts.dnsName, opts.caKeyFile, opts.caCertFile, opts.keyFile, opts.certFile); err != nil {
			cmd.Printf("Failed to generate the TLS key/cert pair: %s\n", err)
			return
		}
		cmd.Printf("TLS key/cert paris gets generated to keyFile: %s, certFile: %s\n", opts.keyFile, opts.certFile)
	}
	return cmd
}
