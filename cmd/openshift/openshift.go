package openshift

import (
	"github.com/gaol/openqe/pkg/openshift"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

const (
	KUBE_CONFIG_ENV = "KUBECONFIG"
)

func BindOcpOptions(opts *openshift.OcpOptions, flags *flag.FlagSet) {
	flags.StringVar(&opts.KUBECONFIG, "kubeconfig", opts.KUBECONFIG, "The kubeconfig file used to communicate with the OpenShift cluster")
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "openshift",
		Short:        "OpenShift oriented test utilities",
		SilenceUsage: true,
	}

	opts := openshift.DefaultOcpOptions()
	BindOcpOptions(opts, cmd.Flags())
	cmd.AddCommand(NewImageRegistryCommand())
	cmd.AddCommand(NewDockerPullSecretCommand())
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	return cmd

}
