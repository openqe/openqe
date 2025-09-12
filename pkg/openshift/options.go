package openshift

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gaol/openqe/pkg/tls"
	"github.com/gaol/openqe/pkg/utils"
	"k8s.io/apimachinery/pkg/util/errors"
)

type OcpOptions struct {
	KUBECONFIG string
}

func DefaultOcpOptions() *OcpOptions {
	ocpOpts := &OcpOptions{}
	if env := os.Getenv("KUBECONFIG"); env != "" {
		ocpOpts.KUBECONFIG = env
	}
	home, err := os.UserHomeDir()
	if err != nil {
		ocpOpts.KUBECONFIG = "./config"
	} else {
		ocpOpts.KUBECONFIG = filepath.Join(home, ".kube", "config")
	}
	return ocpOpts
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

type ImageRegistryOptions struct {
	OcpOpts   *OcpOptions
	PkiOpts   *tls.PKIOptions
	Namespace string
	Image     string
	Name      string
	User      string
	Password  string
	Verbose   bool
}

func DefaultImageRegistryOptions() *ImageRegistryOptions {
	opts := &ImageRegistryOptions{
		OcpOpts:   DefaultOcpOptions(),
		PkiOpts:   tls.DefaultPKIOptions(),
		Namespace: "test-registry",
		Name:      "my-registry",
		Image:     "quay.io/openshifttest/registry:2",
		User:      "reg-user",
		Password:  "reg-pass",
		Verbose:   false,
	}
	return opts
}
