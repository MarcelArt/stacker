/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all services defined in the compose file",
	Long: `List all services defined in the compose file in a tidy table.

Displays each service's name, container name, and published ports.
Services with no published ports show "-" in the PORTS column.

Example:
  stacker -f docker-compose.yml ls`,
	Run: func(cmd *cobra.Command, args []string) {
		headers := []string{"SERVICES", "CONTAINER NAMES", "PORTS"}

		names := make([]string, 0, len(dockerCompose.Services))
		for name := range dockerCompose.Services {
			names = append(names, name)
		}
		sort.Strings(names)

		rows := make([][]string, 0, len(names))
		for _, name := range names {
			svc := dockerCompose.Services[name]
			ports := "-"
			if len(svc.Ports) > 0 {
				ports = strings.Join(svc.Ports, ", ")
			}
			rows = append(rows, []string{name, svc.ContainerName, ports})
		}

		widths := make([]int, len(headers))
		for i, h := range headers {
			widths[i] = len(h)
		}
		for _, row := range rows {
			for i, cell := range row {
				if len(cell) > widths[i] {
					widths[i] = len(cell)
				}
			}
		}

		printRow := func(row []string) {
			for i, cell := range row {
				fmt.Printf("%-*s", widths[i]+3, cell)
			}
			fmt.Println()
		}

		printRow(headers)
		for _, row := range rows {
			printRow(row)
		}
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
