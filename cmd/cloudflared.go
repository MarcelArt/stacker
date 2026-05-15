package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/le-go/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var cloudflaredCmd = &cobra.Command{
	Use:   "cloudflared",
	Short: "Adds cloudflared service",
	Long:  `Adds cloudflared (Cloudflare Tunnel) service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetService("cloudflared", models.Service{
			Image:         "cloudflare/cloudflared:latest",
			ContainerName: "cloudflared",
			Networks:      []string{network},
			Environment: map[string]string{
				"TUNNEL_TOKEN": "${TUNNEL_TOKEN}",
			},
			Command: []string{"tunnel", "run"},
			Restart: "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Cloudflared service successfully added to compose file")
		fmt.Println("Add this line to your .env file:")
		fmt.Println("TUNNEL_TOKEN=<your-tunnel-token>")
	},
}

func init() {
	rootCmd.AddCommand(cloudflaredCmd)
}
