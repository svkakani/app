package main

import (
	"github.com/spf13/cobra"
)

func addCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		pushCmd(),
		pullCmd(),
		listCmd(),
		tagsCmd(),
		renderCmd(),
	)
}

// rootCmd represents the base command when called without any subcommands
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "resto",
		Short:        "Store config files in Docker registries.",
		SilenceUsage: true,
	}
	addCommands(cmd)
	return cmd
}
