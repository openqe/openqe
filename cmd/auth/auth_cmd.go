package auth

import (
	"fmt"

	"github.com/openqe/openqe/pkg/auth"
	"github.com/openqe/openqe/pkg/common"
	"github.com/spf13/cobra"
)

func NewAuthCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "auth",
		Short:         "Authentication related commands",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(NewHtpasswdCommand(globalOpts))
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	return cmd
}

type HtpasswdOption struct {
	username   string
	password   string
	globalOpts *common.GlobalOptions
}

func NewHtpasswdCommand(globalOpts *common.GlobalOptions) *cobra.Command {
	opts := &HtpasswdOption{
		globalOpts: globalOpts,
	}

	cmd := &cobra.Command{
		Use:           "htpasswd",
		Short:         "Create Bcrypt credentials like Apache htpasswd",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().StringVar(&opts.username, "username", opts.username, "The username")
	cmd.Flags().StringVar(&opts.password, "password", opts.password, "The password")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logger := common.NewLoggerFromOptions(opts.globalOpts, "AUTH")

		authCreds, err := auth.GenerateHtpasswdBcrypt(opts.username, opts.password)
		if err != nil {
			return fmt.Errorf("Failed to generate the auth credentials: %w", err)
		}
		logger.Info("htpasswd authentication credentials generated: %s", authCreds)
		return nil
	}
	return cmd
}
