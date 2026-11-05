package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/MarcelArt/stacker/internal/scaffold"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

type options struct {
	file     string
	network  string
	external bool
	compose  scaffold.Document
}

func newRootCommand() *cobra.Command {
	opts := &options{}
	root := &cobra.Command{
		Use:          "stacker",
		Short:        "🍔 stacker - Docker Compose stack builder",
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return cmd.Help() },
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "stacker" {
				return nil
			}
			var config models.Config
			data, err := os.ReadFile("stacker.toml")
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			if err == nil {
				if err := toml.Unmarshal(data, &config); err != nil {
					return fmt.Errorf("read stacker.toml: %w", err)
				}
			}
			if !cmd.Flags().Changed("network") && config.Network != "" {
				opts.network = config.Network
			}
			opts.external = config.IsExternalNetwork
			opts.compose, err = scaffold.ReadCompose(opts.file, cmd.Name() == "add")
			return err
		},
	}
	root.PersistentFlags().StringVarP(&opts.file, "file", "f", "docker-compose.yml", "Target Compose file")
	root.PersistentFlags().StringVarP(&opts.network, "network", "n", "net", "Shared network name (overrides stacker.toml)")
	root.AddCommand(newAddCommand(opts), newListCommand(opts), newRemoveCommand(opts), newConfigCommand(opts))
	return root
}

func Execute() {
	if err := newRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
