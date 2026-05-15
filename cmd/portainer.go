package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var portainerCmd = &cobra.Command{
	Use:   "portainer",
	Short: "Adds portainer service",
	Long:  `Adds portainer service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "portainer_data"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			External: false,
		})
		dockerCompose.SetService("portainer", models.Service{
			Image:         "portainer/portainer-ce:latest",
			ContainerName: "portainer",
			Networks:      []string{network},
			Volumes:       []string{"/var/run/docker.sock:/var/run/docker.sock", fmt.Sprintf("%s:/data", namedVolume)},
			Restart:       "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Portainer service successfully added to compose file")
	},
}

func init() {
	rootCmd.AddCommand(portainerCmd)
}
