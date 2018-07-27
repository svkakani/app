package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/app/pkg/resto"
	"github.com/spf13/cobra"
)

type tagsOptions struct {
	insecure bool
	username string
	password string
}

func tagsCmd() *cobra.Command {
	var opts tagsOptions
	cmd := &cobra.Command{
		Use:   "tags <repository>",
		Short: "List all tags available on a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tags, err := resto.ListRepository(context.Background(), args[0],
				resto.RegistryOptions{
					Username: opts.username,
					Password: opts.password,
					Insecure: opts.insecure,
				})
			if err == nil {
				fmt.Println(strings.Join(tags, "\n"))
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.insecure, "insecure", false, "allow insecure TLS")
	cmd.Flags().StringVar(&opts.username, "user", "", "username to login with")
	cmd.Flags().StringVar(&opts.password, "password", "", "password to login with")
	return cmd
}
