package openshift

import (
	"github.com/openqe/openqe/cmd/core"
	"github.com/openqe/openqe/pkg/openshift"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

func BindImageRegistryOptions(opts *openshift.ImageRegistryOptions, flags *flag.FlagSet) {
	BindOcpOptions(opts.OcpOpts, flags)
	core.BindPKIOptions(opts.PkiOpts, flags)
	flags.StringVar(&opts.Name, "name", opts.Name, "The image registry name")
	flags.StringVar(&opts.Namespace, "namespace", opts.Namespace, "The namespace in which the image registry will be deployed")
	flags.StringVar(&opts.Image, "image", opts.Image, "The image used for the image registry")
	flags.StringVar(&opts.User, "user", opts.User, "The username that can be used to access the image registry")
	flags.StringVar(&opts.Password, "password", opts.Password, "The password that can be used to access the image registry")
	flags.BoolVar(&opts.Verbose, "verbose", opts.Verbose, "If more information should be printed during the setup.")
}

func NewImageRegistryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create-image-registry",
		Short:        "Create an image registry on the current OpenShift cluster with tls and authentication enabled",
		SilenceUsage: true,
	}

	opts := openshift.DefaultImageRegistryOptions()
	BindImageRegistryOptions(opts, cmd.Flags())
	cmd.Run = func(cmd *cobra.Command, args []string) {
		route, err := openshift.SetupImageRegistry(opts, cmd.OutOrStdout())
		if err != nil {
			cmd.Printf("Failed to create the image registry: %s\n", err)
		} else {
			cmd.Printf("Image registry: %s was created.\n", route)
		}
	}
	return cmd
}
