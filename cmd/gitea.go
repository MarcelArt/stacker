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

// giteaCmd represents the gitea command
var giteaCmd = &cobra.Command{
	Use:   "gitea",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetService("gitea", models.Service{
			Image:         "docker.gitea.com/gitea:1.26.2",
			ContainerName: "gitea",
			Ports:         []string{"3000:3000", "222:22"},
			Environment: map[string]string{
				"USER_UID":                 "1000",
				"USER_GID":                 "1000",
				"GITEA__database__DB_TYPE": "postgres",
				"GITEA__database__HOST":    "postgres:5432",
				"GITEA__database__NAME":    "gitea",
				"GITEA__database__USER":    "${POSTGRES_USER}",
				"GITEA__database__PASSWD":  "${POSTGRES_PASSWORD}",
			},
			Restart:  "unless-stopped",
			Networks: []string{network},
			Volumes: []string{
				"./gitea:/data",
				"/etc/timezone:/etc/timezone:ro",
				"/etc/localtime:/etc/localtime:ro",
			},
			DependsOn: []string{"postgres"},
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Gitea service successfully added to compose file")
		fmt.Println("This service requires postgres. Run 'stacker postgres' first if you haven't already.")
	},
}

func init() {
	rootCmd.AddCommand(giteaCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// giteaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// giteaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
