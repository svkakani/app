package main

import (
	"context"
	"fmt"

	"github.com/docker/app/pkg/resto"
	"github.com/spf13/cobra"
)

type pullOptions struct {
	insecure bool
	username string
	password string
}

func pullCmd() *cobra.Command {
	var opts pullOptions
	cmd := &cobra.Command{
		Use:   "pull <repotag>",
		Short: "Pull a file from a registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := resto.PullConfig(context.Background(), args[0],
				resto.RegistryOptions{
					Username: opts.username,
					Password: opts.password,
					Insecure: opts.insecure,
				})
			if err == nil {
				fmt.Printf("%s\n", payload)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.insecure, "insecure", false, "allow insecure TLS")
	cmd.Flags().StringVar(&opts.username, "user", "", "username to login with")
	cmd.Flags().StringVar(&opts.password, "password", "", "password to login with")
	return cmd
}
