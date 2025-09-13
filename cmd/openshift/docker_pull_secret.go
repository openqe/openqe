package openshift

import (
	"github.com/openqe/openqe/pkg/openshift"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

type DockerPullSecretCmdOptions struct {
	OcpOpts    *openshift.OcpOptions
	Namespace  string
	SecretName string
	Auths      []string
	Verbose    bool
}

// BindUpsertDockerPullSecretOptions binds the Docker pull secret options to the command flags
func BindUpsertDockerPullSecretOptions(opts *DockerPullSecretCmdOptions, flags *flag.FlagSet) {
	BindOcpOptions(opts.OcpOpts, flags)
	flags.StringVar(&opts.SecretName, "secret-name", opts.SecretName, "The name of the Docker pull secret")
	flags.StringVar(&opts.Namespace, "namespace", opts.Namespace, "The namespace in which the Docker pull secret will be created")
	flags.BoolVar(&opts.Verbose, "verbose", opts.Verbose, "If more information should be printed during the operation")
	flags.StringArrayVar(&opts.Auths, "auth", nil, "Auth in form <registry>=<username>:<password>[:<email>]. You can specify multiple auths")
}

// NewDockerPullSecretCommand creates the root command for Docker pull secret operations
func NewDockerPullSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "docker-pull-secret",
		Short:        "Docker pull secret management utilities",
		SilenceUsage: true,
	}

	cmd.AddCommand(UpsertDockerPullSecretCommand())
	cmd.AddCommand(NewValidateDockerPullSecretCommand())

	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	return cmd
}

// UpsertDockerPullSecretCommand creates the command for creating a Docker pull secret
func UpsertDockerPullSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "Create or update a Docker pull secret",
		Long:  "Create or update a Docker pull secret in the specified namespace with the provided registry credentials",
	}

	opts := &DockerPullSecretCmdOptions{
		OcpOpts:   openshift.DefaultOcpOptions(),
		Namespace: "default",
	}
	BindUpsertDockerPullSecretOptions(opts, cmd.Flags())
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if opts.SecretName == "" {
			cmd.Printf("Error: --secret-name is required\n")
			return
		}
		if opts.Namespace == "" {
			cmd.Printf("Error: --namespace is required\n")
			return
		}
		if opts.Auths == nil || len(opts.Auths) == 0 {
			cmd.Printf("Error: at least one --auth is required\n")
			return
		}

		dockerPullSecretOpts := openshift.DefaultDockerPullSecretOptions()
		dockerPullSecretOpts.OcpOpts = opts.OcpOpts
		dockerPullSecretOpts.Namespace = opts.Namespace
		dockerPullSecretOpts.SecretName = opts.SecretName
		dockerPullSecretOpts.Verbose = opts.Verbose
		dockerCfg, err := openshift.NewDockerConfig(opts.Auths)
		if err != nil {
			cmd.Printf("Failed to create Docker Config: %s\n", err)
			return
		}
		dockerPullSecretOpts.DockerCfg = dockerCfg
		_, err = openshift.UpsertDockerPullSecret(dockerPullSecretOpts, cmd.OutOrStdout())
		if err != nil {
			cmd.Printf("Failed to create or update Docker pull secret: %s\n", err)
			return
		}
	}
	return cmd
}

type ValidateDockerPullSecretCmdOptions struct {
	ocpOpts        *openshift.OcpOptions
	registryURL    string
	pullSecretFile string
	verbose        bool
}

// NewValidateDockerPullSecretCommand creates the command for validating a Docker pull secret against a registry
func NewValidateDockerPullSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a Docker pull secret",
		Long:  "Validate a Docker pull secret by testing authentication with the registry",
	}

	opts := &ValidateDockerPullSecretCmdOptions{
		ocpOpts: openshift.DefaultOcpOptions(),
	}
	flags := cmd.Flags()
	BindOcpOptions(opts.ocpOpts, flags)
	flags.StringVar(&opts.registryURL, "registry-url", opts.registryURL, "The Docker registry URL to validate against (e.g., quay.io)")
	flags.StringVar(&opts.pullSecretFile, "pull-secret-file", opts.pullSecretFile, "The pull secret file where to configure the authentication")
	flags.BoolVar(&opts.verbose, "verbose", opts.verbose, "If more information should be printed during the operation")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if opts.pullSecretFile == "" {
			cmd.Printf("Error: --pull-secret-file is required\n")
			return
		}
		if opts.registryURL == "" {
			cmd.Printf("Error: --registry-url is required\n")
			return
		}
		valid, err := openshift.ValidateDockerPullSecret(opts.ocpOpts.KUBECONFIG, opts.registryURL, opts.pullSecretFile, opts.verbose)
		if err != nil {
			cmd.Printf("Failed to validate Docker pull secret: %s\n", err)
			return
		}

		if valid {
			cmd.Printf("Pull secret file: %s is valid for registry %s\n", opts.pullSecretFile, opts.registryURL)
		} else {
			cmd.Printf("Pull secret file: %s is NOT valid for registry %s\n", opts.pullSecretFile, opts.registryURL)
		}
	}
	return cmd
}
