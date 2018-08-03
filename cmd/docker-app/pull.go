package main

import (
	"github.com/docker/app/lib/dockerapp"
	"github.com/docker/app/internal"
	"github.com/docker/cli/cli"
	"github.com/spf13/cobra"
)

func pullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull <repotag>",
		Short: "Pull an application from a registry",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := dockerapp.FromImage(args[0])
			if err != nil {
				return err
			}
			return dockerapp.ToDirectory(app, internal.DirNameFromAppName(app.AppName))
		},
	}
}
