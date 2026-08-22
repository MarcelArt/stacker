package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var infisicalCmd = &cobra.Command{
	Use:   "infisical",
	Short: "Adds infisical service",
	Long:  `Adds infisical service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: isExternalNetwork,
		})
		dockerCompose.SetService("infisical", models.Service{
			Image:         "infisical/infisical:latest",
			ContainerName: "infisical",
			Restart:       "unless-stopped",
			DependsOn:     []string{"postgres", "redis"},
			PullPolicy:    "always",
			EnvFile:       ".env.infisical",
			Ports:         []string{"80:8080"},
			Environment: map[string]string{
				"NODE_ENV": "production",
			},
			Networks: []string{network},
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Infisical service successfully added to compose file")
		fmt.Println("This service requires postgres and redis. Run 'stacker postgres' and 'stacker redis' first if you haven't already.")
		fmt.Println("Please run this command to generate .env.infisical file:")
		fmt.Println("curl -o .env.infisical https://raw.githubusercontent.com/Infisical/infisical/main/.env.example")
	},
}

func init() {
	rootCmd.AddCommand(infisicalCmd)
}
