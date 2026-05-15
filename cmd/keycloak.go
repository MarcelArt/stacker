package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var keycloakCmd = &cobra.Command{
	Use:   "keycloak",
	Short: "Adds keycloak service",
	Long:  `Adds keycloak service to your compose file`,
	Run: func(cmd *cobra.Command, args []string) {
		dockerCompose.SetNetwork(network, models.Network{
			Name:     network,
			Driver:   "bridge",
			External: false,
		})
		dockerCompose.SetService("keycloak", models.Service{
			Image:         "quay.io/keycloak/keycloak:latest",
			ContainerName: "keycloak",
			Networks:      []string{network},
			Environment: map[string]string{
				"KC_DB":                      "postgres",
				"KC_DB_URL":                  "jdbc:postgresql://postgres:5432/keycloak",
				"KC_DB_USERNAME":             "${POSTGRES_USER}",
				"KC_DB_PASSWORD":             "${POSTGRES_PASSWORD}",
				"KEYCLOAK_ADMIN":             "${KEYCLOAK_ADMIN}",
				"KEYCLOAK_ADMIN_PASSWORD":    "${KEYCLOAK_ADMIN_PASSWORD}",
				"KC_HOSTNAME":                "${KC_HOSTNAME}",
				"KC_HOSTNAME_STRICT":         "false",
				"KC_HOSTNAME_STRICT_HTTPS":   "true",
				"KC_HTTP_ENABLED":            "true",
				"KC_PROXY_HEADERS":           "xforwarded",
				"KC_PROXY_TRUSTED_ADDRESSES": "127.0.0.0/8",
			},
			Command:   []string{"start"},
			Ports:     []string{"9080:8080", "8443:8443"},
			DependsOn: []string{"postgres"},
			Restart:   "unless-stopped",
		})

		yml, err := yaml.Marshal(dockerCompose)
		if err != nil {
			log.Fatalf("failed serialize to yml: %v", err.Error())
		}
		if err := os.WriteFile(composeFile, yml, 0644); err != nil {
			log.Fatalf("failed writing compose file: %v", err.Error())
		}

		fmt.Println("Keycloak service successfully added to compose file")
		fmt.Println("This service requires postgres. Run 'le-go postgres' first if you haven't already.")
		fmt.Println("Add these lines to your .env file:")
		fmt.Println("KEYCLOAK_ADMIN=<your-keycloak-admin>")
		fmt.Println("KEYCLOAK_ADMIN_PASSWORD=<your-keycloak-admin-password>")
		fmt.Println("KC_HOSTNAME=<your-keycloak-hostname>")
	},
}

func init() {
	rootCmd.AddCommand(keycloakCmd)
}
