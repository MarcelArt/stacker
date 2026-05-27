package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var jenkinsCmd = &cobra.Command{
	Use:   "jenkins",
	Short: "Adds jenkins service",
	Long:  `Adds jenkins service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "jenkins_home"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: isExternalNetwork,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			External: isExternalNetwork,
		})
		dockerCompose.SetService("jenkins", models.Service{
			Image:         "jenkins/jenkins:lts",
			ContainerName: "jenkins",
			Networks:      []string{network},
			Ports:         []string{"8080:8080", "50000:50000"},
			Volumes:       []string{fmt.Sprintf("%s:/var/jenkins_home", namedVolume), "/var/run/docker.sock:/var/run/docker.sock"},
			User:          "root",
			Restart:       "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Jenkins service successfully added to compose file")
	},
}

func init() {
	rootCmd.AddCommand(jenkinsCmd)
}
