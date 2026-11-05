package cmd

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/MarcelArt/stacker/internal/scaffold"
	"github.com/spf13/cobra"
)

func newListCommand(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List services in the Compose file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			table := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(table, "SERVICES\tCONTAINER NAMES\tPORTS")
			services := opts.compose.Section("services")
			names := make([]string, 0, len(services))
			for name := range services {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				service := services[name].(map[string]any)
				container, _ := service["container_name"].(string)
				if container == "" {
					container = "-"
				}
				fmt.Fprintf(table, "%s\t%s\t%s\n", name, container, scaffold.PortSummary(service))
			}
			return table.Flush()
		},
	}
}
