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

// natsCmd represents the nats command
var natsCmd = &cobra.Command{
	Use:   "nats",
	Short: "Adds nats service",
	Long:  `Adds nats service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "nats"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: isExternalNetwork,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			External: isExternalNetwork,
		})
		dockerCompose.SetService("nats", models.Service{
			Image:         "nats:latest",
			ContainerName: "nats",
			Networks:      []string{network},
			Ports:         []string{"4222:4222", "8222:8222"},
			Volumes:       []string{fmt.Sprintf("%s:/data", namedVolume)},
			Command:       "-js -sd /data -m 8222",
			Restart:       "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Nats service successfully added to compose file")
	},
}

func init() {
	rootCmd.AddCommand(natsCmd)
}
