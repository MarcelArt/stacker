package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var redisCmd = &cobra.Command{
	Use:   "redis",
	Short: "Adds redis service",
	Long:  `Adds redis service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		namedVolume := "redis_data"
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: isExternalNetwork,
		})
		dockerCompose.SetVolume(namedVolume, models.Volume{
			Name:     namedVolume,
			Driver:   "local",
			External: isExternalNetwork,
		})
		dockerCompose.SetService("redis", models.Service{
			Image:         "redis",
			ContainerName: "redis",
			EnvFile:       []string{".env"},
			Restart:       "always",
			Environment: map[string]string{
				"ALLOW_EMPTY_PASSWORD": "yes",
			},
			Networks: []string{network},
			Volumes:  []string{fmt.Sprintf("%s:/data", namedVolume)},
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Redis service successfully added to compose file")
	},
}

func init() {
	rootCmd.AddCommand(redisCmd)
}
