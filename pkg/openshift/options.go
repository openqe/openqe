package openshift

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// DockerAuthEntry represents an entry in the Docker config.json auths
type DockerAuthEntry struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth"`
}

// DockerConfig represents the structure of a Docker config.json file
type DockerConfig struct {
	Auths map[string]DockerAuthEntry `json:"auths"`
}

// DockerPullSecretOptions contains the options for creating a Docker pull secret
type DockerPullSecretOptions struct {
	OcpOpts    *OcpOptions
	DockerCfg  *DockerConfig
	Namespace  string
	SecretName string
	Verbose    bool
}

func DefaultDockerPullSecretOptions() *DockerPullSecretOptions {
	return &DockerPullSecretOptions{
		OcpOpts: DefaultOcpOptions(),
		DockerCfg: &DockerConfig{
			Auths: map[string]DockerAuthEntry{},
		},
		Namespace: "default",
	}
}

// NewDockerConfig creates a DockerConfig with input string slice: <auth>=<user>:<pass>[:<email>]
// It does not support ':' in the username
func NewDockerConfig(auths []string) (*DockerConfig, error) {
	dockerCfg := &DockerConfig{
		Auths: map[string]DockerAuthEntry{},
	}
	for _, a := range auths {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid auth format: %s", a)
		}
		auth := parts[0]
		rest := parts[1] // username:password:email
		// Split into fields, but allow `:` in the password by taking the last one(s) from the right
		fields := strings.Split(rest, ":")
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid format (need at least user:pass): %s", rest)
		}
		username := fields[0]
		email := ""
		password := ""
		// If 3+ fields → last one is email, everything between username and email is password
		if len(fields) >= 3 {
			email = fields[len(fields)-1]
			password = strings.Join(fields[1:len(fields)-1], ":") // rejoin with `:` for passwords
		} else {
			password = fields[1]
		}
		dockerCfg.Auths[auth] = DockerAuthEntry{
			Username: username,
			Password: password,
			Email:    email,
			Auth:     base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))),
		}
	}
	return dockerCfg, nil
}

// MergeDockerConfig merges the DockerConfig represented by cfg to the DockerConfig represented by target and returns the merged DockerConfig
func MergeDockerConfig(target, cfg *DockerConfig) *DockerConfig {
	finalCfg := &DockerConfig{
		Auths: target.Auths,
	}
	for k, v := range cfg.Auths {
		finalCfg.Auths[k] = v
	}
	return finalCfg
}
