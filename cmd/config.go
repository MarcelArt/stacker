package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MarcelArt/stacker/v2/internal/models"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

func newConfigCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Save shared network defaults from the Compose file to stacker.toml",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			networks := opts.compose.Section("networks")
			selected := opts.network
			if _, present := networks[selected]; !present {
				if cmd.Flags().Changed("network") {
					return fmt.Errorf("network %q is not defined in %s", selected, opts.file)
				}
				if len(networks) != 1 {
					names := make([]string, 0, len(networks))
					for name := range networks {
						names = append(names, name)
					}
					sort.Strings(names)
					return fmt.Errorf("select a defined network with -n (available: %s)", strings.Join(names, ", "))
				}
				for name := range networks {
					selected = name
				}
			}
			external := false
			if definition, ok := networks[selected].(map[string]any); ok {
				external, _ = definition["external"].(bool)
			}
			config := models.Config{Network: selected, IsExternalNetwork: external}
			data, err := toml.Marshal(config)
			if err != nil {
				return err
			}
			if err := os.WriteFile("stacker.toml", data, 0644); err != nil {
				return err
			}
			cmd.Printf("Created stacker.toml using network %s from %s\n", selected, opts.file)
			return nil
		},
	}
}
