package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var nakamaCmd = &cobra.Command{
	Use:   "nakama",
	Short: "Adds nakama service",
	Long:  `Adds nakama game server service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetService("nakama", models.Service{
			Image:         "registry.heroiclabs.com/heroiclabs/nakama:3.22.0",
			ContainerName: "nakama",
			Networks:      []string{network},
			Entrypoint: []string{
				"/bin/sh",
				"-ecx",
				"/nakama/nakama migrate up --database.address ${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/nakama && exec /nakama/nakama --name nakama1 --database.address ${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/nakama --logger.level DEBUG --session.token_expiry_sec 7200",
			},
			Expose: []string{"7349", "7350", "7351"},
			Ports:  []string{"7349:7349", "7350:7350", "7351:7351"},
			Volumes: []string{
				"./nakama/data:/nakama/data",
			},
			Healthcheck: &models.Healthcheck{
				Test:     []string{"CMD", "/nakama/nakama", "healthcheck"},
				Interval: "10s",
				Timeout:  "5s",
				Retries:  5,
			},
			Links:     []string{"postgres:db"},
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

		fmt.Println("Nakama service successfully added to compose file")
		fmt.Println("This service requires postgres. Run 'le-go postgres' first if you haven't already.")
	},
}

func init() {
	rootCmd.AddCommand(nakamaCmd)
}
