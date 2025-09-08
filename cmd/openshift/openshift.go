package openshift

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/errors"

	"github.com/gaol/openqe/pkg/utils"
	"github.com/spf13/cobra"
)

const (
	KUBE_CONFIG_ENV = "KUBECONFIG"
)

type OcpOptions struct {
	KUBECONFIG string
}

func (o *OcpOptions) Validate() error {
	var errs []error
	if o.KUBECONFIG == "" {
		errs = append(errs, fmt.Errorf("--kubeconfig must be specified"))
	} else if !utils.FileExists(o.KUBECONFIG) {
		errs = append(errs, fmt.Errorf("--kubeconfig %s does not exist", o.KUBECONFIG))
	}

	return errors.NewAggregate(errs)
}

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "openshift",
		Short:        "OpenShift oriented test utilities",
		SilenceUsage: true,
	}

	var opts OcpOptions
	opts.KUBECONFIG = "~/.kube/config"

	cmd.PersistentFlags().StringVar(&opts.KUBECONFIG, "kubeconfig", "~/.kube/config", "The kubeconfig file used to communicate with the OpenShift cluster")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	return cmd

}
