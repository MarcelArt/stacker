package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/le-go/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var grafanaCmd = &cobra.Command{
	Use:   "grafana",
	Short: "Adds grafana service",
	Long:  `Adds grafana (OTel LGTM) service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetVolume("loki_data", models.Volume{
			Name:     "loki_data",
			External: false,
		})
		dockerCompose.SetVolume("grafana_data", models.Volume{
			Name:     "grafana_data",
			External: false,
		})
		dockerCompose.SetVolume("prometheus_data", models.Volume{
			Name:     "prometheus_data",
			External: false,
		})
		dockerCompose.SetService("grafana", models.Service{
			Image:         "grafana/otel-lgtm:latest",
			ContainerName: "grafana",
			Ports:         []string{"3000:3000", "4317:4317", "4318:4318", "3100:3100"},
			Networks:      []string{network},
			Volumes:       []string{"loki_data:/loki", "grafana_data:/var/lib/grafana", "prometheus_data:/prometheus"},
			Restart:       "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Grafana service successfully added to compose file")
	},
}

func init() {
	rootCmd.AddCommand(grafanaCmd)
}
