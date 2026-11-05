package cmd

import (
	"fmt"

	"github.com/MarcelArt/stacker/internal/scaffold"
	"github.com/spf13/cobra"
)

func newRemoveCommand(opts *options) *cobra.Command {
	var container string
	cmd := &cobra.Command{
		Use:   "rm -c <service>",
		Short: "Remove a service while preserving unrelated Compose fields",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if container == "" {
				return fmt.Errorf("provide a service name with --container or -c")
			}
			services := opts.compose.Section("services")
			if _, present := services[container]; !present {
				return fmt.Errorf("service %q does not exist", container)
			}
			delete(services, container)
			if err := scaffold.WriteCompose(opts.file, opts.compose); err != nil {
				return err
			}
			cmd.Printf("%s service removed\n", container)
			return nil
		},
	}
	cmd.Flags().StringVarP(&container, "container", "c", "", "Service name to remove")
	return cmd
}
