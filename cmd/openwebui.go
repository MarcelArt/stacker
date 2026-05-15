package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/le-go/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var openWebUICmd = &cobra.Command{
	Use:   "open-web-ui",
	Short: "Adds open-web-ui service",
	Long:  `Adds Open Web UI service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "open-webui"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			External: false,
		})
		dockerCompose.SetService("open-web-ui", models.Service{
			Image:         "ghcr.io/open-webui/open-webui:main",
			ContainerName: "open-web-ui",
			Networks:      []string{network},
			Ports:         []string{"3000:8080"},
			Environment: map[string]string{
				"OLLAMA_BASE_URL": "http://ollama:11434",
			},
			Volumes:   []string{fmt.Sprintf("%s:/app/backend/data", namedVolume)},
			DependsOn: []string{"ollama"},
			Restart:   "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Open Web UI service successfully added to compose file")
		fmt.Println("This service requires ollama. Run 'le-go ollama' first if you haven't already.")
	},
}

func init() {
	rootCmd.AddCommand(openWebUICmd)
}
