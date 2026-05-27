/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var defaultNetwork string

		i := 0
		for k := range dockerCompose.Networks {
			if i > 0 {
				break
			}

			defaultNetwork = k
			i++
		}

		stackerConfig := models.Config{Network: defaultNetwork}

		stackerTOML, err := toml.Marshal(stackerConfig)
		if err != nil {
			log.Fatalf("failed to serialize toml: %s", err.Error())
		}

		if err := os.WriteFile("stacker.toml", stackerTOML, 0644); err != nil {
			log.Fatalf("failed to write toml: %s", err.Error())
		}

		fmt.Println("Created stacker.toml file")
		fmt.Printf("Default configuration imported from %s\n", composeFile)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
