package openshift

import (
	"fmt"

	"github.com/openqe/openqe/cmd/core"
	"github.com/openqe/openqe/pkg/common"
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
}

func NewImageRegistryCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create-image-registry",
		Short:        "Create an image registry on the current OpenShift cluster with tls and authentication enabled",
		SilenceUsage: true,
	}

	opts := openshift.DefaultImageRegistryOptions()
	opts.GlobalOpts = globalOpts
	BindImageRegistryOptions(opts, cmd.Flags())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logger := common.NewLoggerFromOptions(globalOpts, "OPENSHIFT")

		route, err := openshift.SetupImageRegistry(opts)
		if err != nil {
			return fmt.Errorf("Failed to create the image registry: %v", err)
		}
		logger.Info("Image registry: %s was created.", route)
		return nil
	}
	return cmd
}
