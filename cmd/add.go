package cmd

import (
	"sort"
	"strings"

	"github.com/MarcelArt/stacker/v2/internal/scaffold"
	"github.com/spf13/cobra"
)

func newAddCommand(opts *options) *cobra.Command {
	var overwrite bool
	cmd := &cobra.Command{
		Use:   "add <template-name|https-url>",
		Short: "Add services from a built-in or HTTPS YAML template",
		Long:  "Add services, networks, and volumes from a Compose-compatible YAML template.\n\nBuilt-in templates: " + strings.Join(scaffold.Builtins(), ", "),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			template, err := scaffold.LoadTemplate(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if err := template.BindNetwork(opts.network, opts.external); err != nil {
				return err
			}
			merged, err := scaffold.Merge(opts.compose, template.Compose, overwrite)
			if err != nil {
				return err
			}
			envPath := scaffold.EnvironmentPath(opts.file)
			env, pending, err := scaffold.PrepareEnvironment(envPath, template.Metadata.Environment)
			if err != nil {
				return err
			}
			if err := scaffold.WriteStack(opts.file, merged, env); err != nil {
				return err
			}
			var names []string
			for name := range template.Compose.Section("services") {
				names = append(names, name)
			}
			sort.Strings(names)
			cmd.Printf("Added %s to %s\n", strings.Join(names, ", "), opts.file)
			if template.Metadata.Description != "" {
				cmd.Println(template.Metadata.Description)
			}
			var variables []string
			for name := range template.Metadata.Environment {
				variables = append(variables, name)
			}
			sort.Strings(variables)
			for _, name := range variables {
				cmd.Printf("%s: %s\n", name, template.Metadata.Environment[name].Description)
			}
			if len(pending) > 0 {
				cmd.Printf("Fill in %s: %s\n", envPath, strings.Join(pending, ", "))
			}
			for _, dependency := range scaffold.MissingDependencies(merged, template.Compose) {
				cmd.Printf("Missing dependency %s: add it with `stacker add %s` or a template containing that service.\n", dependency, dependency)
			}
			for _, instruction := range template.Metadata.Instructions {
				cmd.Println(instruction)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "Replace existing services (resource conflicts still fail)")
	cmd.ValidArgs = scaffold.Builtins()
	return cmd
}
