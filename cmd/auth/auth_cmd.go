package auth

import (
	"github.com/openqe/openqe/pkg/auth"
	"github.com/spf13/cobra"
)

func NewAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "auth",
		Short:         "Authentication related commands",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(NewHtpasswdCommand())
	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmd.Help()
	}
	return cmd
}

type HtpasswdOption struct {
	username string
	password string
}

func NewHtpasswdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "htpasswd",
		Short:         "Create Bcrypt credentials like Apache htpasswd",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	opts := &HtpasswdOption{}
	cmd.Flags().StringVar(&opts.username, "username", opts.username, "The username")
	cmd.Flags().StringVar(&opts.password, "password", opts.password, "The password")
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		auth, err := auth.GenerateHtpasswdBcrypt(opts.username, opts.password)
		if err != nil {
			cmd.Printf("Failed to generate the auth credentials: %s\n", err)
			return
		}
		cmd.Printf("%s\n", auth)
	}
	return cmd
}
