package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/app/pkg/resto"
	"github.com/spf13/cobra"
)

type listOptions struct {
	insecure bool
	username string
	password string
}

func listCmd() *cobra.Command {
	var opts listOptions
	cmd := &cobra.Command{
		Use:   "list <registry>",
		Short: "List repositories from a registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := resto.ListRegistry(context.Background(), args[0],
				resto.RegistryOptions{
					Username: opts.username,
					Password: opts.password,
					Insecure: opts.insecure,
				})
			if err == nil {
				fmt.Println(strings.Join(repos, "\n"))
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.insecure, "insecure", false, "allow insecure TLS")
	cmd.Flags().StringVar(&opts.username, "user", "", "username to login with")
	cmd.Flags().StringVar(&opts.password, "password", "", "password to login with")
	return cmd
}
