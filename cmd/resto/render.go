package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docker/app/internal/settings"
	"github.com/docker/app/internal/yatee"
	"github.com/docker/app/pkg/resto"
	cliopts "github.com/docker/cli/opts"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var (
	renderSettingsFile []string
	renderEnv          []string
	renderOutput       string
)

func renderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "render <yaml-file> [-s key=value...] [-f settings-file...]",
		Short: "Render given YAML file",
		Long:  `Render given YAML file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := cliopts.ConvertKVStringsToMap(renderEnv)
			envSettings, err := settings.FromFlatten(d)
			if err != nil {
				return err
			}
			s, err := yatee.LoadSettings(renderSettingsFile)
			if err != nil {
				return err
			}
			allSettings, err := settings.Merge(s, envSettings)
			if err != nil {
				return err
			}
			input, err := ioutil.ReadFile(args[0])
			if err != nil {
				payload, err := resto.PullConfig(context.Background(), args[0], resto.RegistryOptions{Insecure: os.Getenv("DOCKERAPP_INSECURE_REGISTRY") != ""})
				if err != nil {
					return fmt.Errorf("failed to find input on disk or registry")
				}
				input = []byte(payload)
			}
			result, err := yatee.Process(string(input), allSettings)
			if err != nil {
				return err
			}
			resultBytes, err := yaml.Marshal(result)
			if err != nil {
				return err
			}
			if renderOutput == "-" {
				os.Stdout.Write(resultBytes)
			}
			return ioutil.WriteFile(renderOutput, resultBytes, 0644)
		},
	}
	cmd.Flags().StringArrayVarP(&renderSettingsFile, "settings-files", "f", []string{}, "Override settings files")
	cmd.Flags().StringArrayVarP(&renderEnv, "set", "s", []string{}, "Override settings values")
	cmd.Flags().StringVarP(&renderOutput, "output", "o", "-", "Output file")
	return cmd
}
