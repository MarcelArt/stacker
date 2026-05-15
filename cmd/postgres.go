/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/le-go/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// postgresCmd represents the postgres command
var postgresCmd = &cobra.Command{
	Use:   "postgres",
	Short: "Adds postgres service",
	Long:  `Adds postgres service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "postgres_data"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			External: false,
		})
		dockerCompose.SetService("postgres", models.Service{
			Image:         "postgres:17",
			ContainerName: "postgres",
			Ports:         []string{"5432:5432"},
			Environment: map[string]string{
				"POSTGRES_USER":     "${POSTGRES_USER}",
				"POSTGRES_PASSWORD": "${POSTGRES_PASSWORD}",
			},
			Restart:  "unless-stopped",
			Networks: []string{network},
			Volumes:  []string{fmt.Sprintf("%s:/var/lib/postgresql/data", namedVolume)},
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Postgres service successfully added to compose file")
		fmt.Println("Add this line to your .env file:")
		fmt.Println("POSTGRES_USER=<your-pg-user>")
		fmt.Println("POSTGRES_PASSWORD=<your-pg-password>")
	},
}

func init() {
	rootCmd.AddCommand(postgresCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// postgresCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// postgresCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
