package main

import (
	"os"

	"github.com/docker/app/lib/dockerapp"
	"github.com/docker/cli/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var splitOutputDir string

func splitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "split [<app-name>] [-o output]",
		Short: "Split a single-file application into multiple files",
		Args:  cli.RequiresMaxArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := dockerapp.Load(firstOrEmpty(args))
			if err != nil {
				return err
			}
			inPlace := splitOutputDir == ""
			if inPlace {
				splitOutputDir = app.Origin + ".tmp"
			}
			if err := dockerapp.ToDirectory(app, splitOutputDir); err != nil {
				return err
			}
			if inPlace {
				if err := os.RemoveAll(app.Origin); err != nil {
					return errors.Wrap(err, "failed to erase previous application directory")
				}
				if err := os.Rename(splitOutputDir, app.Origin); err != nil {
					return errors.Wrap(err, "failed to rename new application directory")
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&splitOutputDir, "output", "o", "", "Output application directory (default: in-place)")
	return cmd
}
