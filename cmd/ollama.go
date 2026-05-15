package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var ollamaCmd = &cobra.Command{
	Use:   "ollama",
	Short: "Adds ollama service",
	Long:  `Adds ollama LLM service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "ollama_data"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			External: false,
		})
		dockerCompose.SetService("ollama", models.Service{
			Image:         "ollama/ollama:latest",
			ContainerName: "ollama",
			Networks:      []string{network},
			Ports:         []string{"11434:11434"},
			Volumes:       []string{fmt.Sprintf("%s:/root/.ollama", namedVolume)},
			Restart:       "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Ollama service successfully added to compose file")
	},
}

func init() {
	rootCmd.AddCommand(ollamaCmd)
}
