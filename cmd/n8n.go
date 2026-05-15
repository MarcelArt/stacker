package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/le-go/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var n8nCmd = &cobra.Command{
	Use:   "n8n",
	Short: "Adds n8n service",
	Long:  `Adds n8n workflow automation service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "n8n_data"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			External: false,
		})
		dockerCompose.SetService("n8n", models.Service{
			Image:         "docker.n8n.io/n8nio/n8n",
			ContainerName: "n8n",
			Networks:      []string{network},
			Ports:         []string{"5678:5678"},
			Volumes:       []string{fmt.Sprintf("%s:/home/node/.n8n", namedVolume)},
			Environment: map[string]string{
				"DB_TYPE":                "postgresdb",
				"DB_POSTGRESDB_DATABASE": "n8n",
				"DB_POSTGRESDB_HOST":     "postgres",
				"DB_POSTGRESDB_PORT":     "5432",
				"DB_POSTGRESDB_USER":     "${POSTGRES_USER}",
				"DB_POSTGRESDB_SCHEMA":   "public",
				"DB_POSTGRESDB_PASSWORD": "${POSTGRES_PASSWORD}",
			},
			DependsOn: []string{"postgres"},
			Restart:   "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("n8n service successfully added to compose file")
		fmt.Println("This service requires postgres. Run 'le-go postgres' first if you haven't already.")
	},
}

func init() {
	rootCmd.AddCommand(n8nCmd)
}
