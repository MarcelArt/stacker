/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/MarcelArt/stacker/internal/models"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var composeFile string
var dockerCompose models.Compose
var network string
var isExternalNetwork = false

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stacker",
	Short: "🍔 stacker - Docker Compose stack builder",
	Long:  `🍔 Stack up your Docker Compose infrastructure one service at a time.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		content, _ := os.ReadFile(composeFile)
		yaml.Unmarshal(content, &dockerCompose)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	var stackerConfig models.Config
	stackerTOML, _ := os.ReadFile("stacker.toml")
	toml.Unmarshal(stackerTOML, &stackerConfig)

	defaultNetwork := "net"
	if stackerConfig.Network != "" {
		defaultNetwork = stackerConfig.Network
		isExternalNetwork = stackerConfig.IsExternalNetwork
	}

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.stacker.yaml)")
	rootCmd.PersistentFlags().StringVarP(&composeFile, "file", "f", "docker-compose.yml", "Compose file")
	rootCmd.PersistentFlags().StringVarP(&network, "network", "n", defaultNetwork, "Network name")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
