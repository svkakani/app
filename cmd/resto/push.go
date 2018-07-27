package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/docker/app/pkg/resto"
	"github.com/spf13/cobra"
)

type pushOptions struct {
	insecure bool
	username string
	password string
}

func pushCmd() *cobra.Command {
	var opts pushOptions
	cmd := &cobra.Command{
		Use:   "push <file> <repotag>",
		Short: "Push a file to a registry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			dgst, err := resto.PushConfig(context.Background(), string(payload), args[1],
				resto.RegistryOptions{
					Username: opts.username,
					Password: opts.password,
					Insecure: opts.insecure,
				}, nil)
			if err == nil {
				fmt.Printf("%v\n", dgst)
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.insecure, "insecure", false, "allow insecure TLS")
	cmd.Flags().StringVar(&opts.username, "user", "", "username to login with")
	cmd.Flags().StringVar(&opts.password, "password", "", "password to login with")
	return cmd
}
