package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var surrealdbCmd = &cobra.Command{
	Use:   "surrealdb",
	Short: "Adds surrealdb service",
	Long:  `Adds surrealdb service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetService("surrealdb", models.Service{
			Image:         "surrealdb/surrealdb:latest",
			ContainerName: "surrealdb",
			Networks:      []string{network},
			Ports:         []string{"8000:8000"},
			Command:       "start --user ${SURREAL_USER} --pass ${SURREAL_PASS} rocksdb:/mydata/surreal.db",
			Volumes:       []string{"./surrealdb:/mydata"},
			User:          "1000:1000",
			Restart:       "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("SurrealDB service successfully added to compose file")
		fmt.Println("Add these lines to your .env file:")
		fmt.Println("SURREAL_USER=<your-surreal-user>")
		fmt.Println("SURREAL_PASS=<your-surreal-password>")
	},
}

func init() {
	rootCmd.AddCommand(surrealdbCmd)
}
