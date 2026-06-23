/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// stirlingCmd represents the stirling command
var stirlingCmd = &cobra.Command{
	Use:   "stirling",
	Short: "Adds stirling-pdf service",
	Long:  "Adds stirling-pdf the pdf processor service to your compose file.",
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: isExternalNetwork,
		})
		dockerCompose.SetService("stirling-pdf", models.Service{
			Image:         "stirlingtools/stirling-pdf:latest",
			ContainerName: "stirling-pdf",
			Networks:      []string{network},
			Ports:         []string{"8080:8080"},
			Volumes:       []string{"./stirling-data:/configs"},
			Restart:       "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("stirling-pdf service successfully added to compose file")
	},
}

func init() {
	rootCmd.AddCommand(stirlingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stirlingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stirlingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
