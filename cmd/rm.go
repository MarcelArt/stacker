/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var container string

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove a service from the compose file",
	Long: `Remove a service from the compose file.

Deletes the service specified by the --container flag and writes the
updated configuration back to the compose file.

Example:
  stacker rm -c postgres`,
	Run: func(cmd *cobra.Command, args []string) {
		if container == "" {
			fmt.Println("Please provide container name by using --container or -c flag, example: stacker rm -c postgres")
			os.Exit(1)
			return
		}

		dockerCompose.DeleteService(container)

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}
		fmt.Printf("%s service removed", container)
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// rmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	rmCmd.Flags().StringVarP(&container, "container", "c", "", "container name to remove")
}
