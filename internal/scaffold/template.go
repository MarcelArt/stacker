package scaffold

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/MarcelArt/stacker/templates"
	"gopkg.in/yaml.v3"
)

const maxTemplateBytes = 1024 * 1024

type EnvironmentRequirement struct {
	Description string `yaml:"description"`
}

type Metadata struct {
	Description   string                            `yaml:"description"`
	SharedNetwork string                            `yaml:"shared_network"`
	Environment   map[string]EnvironmentRequirement `yaml:"environment"`
	Instructions  []string                          `yaml:"instructions"`
}

type Template struct {
	Compose  Document
	Metadata Metadata
}

func Builtins() []string {
	entries, _ := templates.Files.ReadDir(".")
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, strings.TrimSuffix(entry.Name(), ".yml"))
	}
	return names
}

func LoadTemplate(ctx context.Context, source string) (*Template, error) {
	if !strings.ContainsAny(source, "/\\:.") {
		data, err := templates.Files.ReadFile(source + ".yml")
		if err != nil {
			return nil, fmt.Errorf("unknown built-in template %q; available: %s", source, strings.Join(Builtins(), ", "))
		}
		return ParseTemplate(data)
	}
	client := &http.Client{Timeout: 20 * time.Second}
	return loadRemote(ctx, source, client)
}

func loadRemote(ctx context.Context, source string, client *http.Client) (*Template, error) {
	parsed, err := url.Parse(source)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return nil, fmt.Errorf("template source must be a built-in name or an HTTPS URL to raw YAML")
	}
	copyClient := *client
	copyClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if req.URL.Scheme != "https" || req.URL.User != nil {
			return fmt.Errorf("template redirects must use HTTPS without embedded credentials")
		}
		if len(via) >= 5 {
			return fmt.Errorf("too many template redirects")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, err
	}
	response, err := copyClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch template: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch template: HTTP %d; use a raw YAML URL", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxTemplateBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}
	if len(data) > maxTemplateBytes {
		return nil, fmt.Errorf("template exceeds 1 MiB")
	}
	return ParseTemplate(data)
}

var environmentName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var interpolation = regexp.MustCompile(`\$\$|\$\{([A-Za-z_][A-Za-z0-9_]*)|\$([A-Za-z_][A-Za-z0-9_]*)`)

func ParseTemplate(data []byte) (*Template, error) {
	doc, err := ParseCompose(data)
	if err != nil {
		return nil, err
	}
	if len(doc.Section("services")) == 0 {
		return nil, fmt.Errorf("template must define at least one service")
	}
	template := &Template{Compose: doc}
	if raw, present := doc["x-stacker"]; present {
		encoded, err := yaml.Marshal(raw)
		if err != nil {
			return nil, err
		}
		decoder := yaml.NewDecoder(bytes.NewReader(encoded))
		decoder.KnownFields(true)
		if err := decoder.Decode(&template.Metadata); err != nil {
			return nil, fmt.Errorf("invalid x-stacker metadata: %w", err)
		}
		delete(doc, "x-stacker")
	}
	for name, requirement := range template.Metadata.Environment {
		if !environmentName.MatchString(name) || strings.TrimSpace(requirement.Description) == "" {
			return nil, fmt.Errorf("environment variable %q needs a valid name and a description", name)
		}
	}
	if err := validateEnvironment(doc, template.Metadata.Environment); err != nil {
		return nil, err
	}
	if err := validateNetworking(doc); err != nil {
		return nil, err
	}
	return template, nil
}

// Compose interpolates values throughout the document, including entrypoints.
// Escaped dollars ($$) are passed through and do not declare a Compose input.
func validateEnvironment(value any, requirements map[string]EnvironmentRequirement) error {
	switch value := value.(type) {
	case Document:
		return validateEnvironment(map[string]any(value), requirements)
	case map[string]any:
		for _, key := range sortedKeys(value) {
			if err := validateEnvironment(value[key], requirements); err != nil {
				return err
			}
		}
	case []any:
		for _, entry := range value {
			if err := validateEnvironment(entry, requirements); err != nil {
				return err
			}
		}
	case string:
		for _, match := range interpolation.FindAllStringSubmatch(value, -1) {
			name := match[1] + match[2]
			if name == "" {
				continue
			}
			if _, declared := requirements[name]; !declared {
				return fmt.Errorf("variable %s needs a description in x-stacker.environment", name)
			}
		}
	}
	return nil
}

// BindNetwork renames only explicitly declared shared-network references. Services
// using network_mode are left untouched and receive no network attachments.
func (t *Template) BindNetwork(name string, external bool) error {
	shared := t.Metadata.SharedNetwork
	if shared == "" {
		return nil
	}
	networks := t.Compose.Section("networks")
	definition, present := networks[shared]
	if !present {
		return fmt.Errorf("shared_network %q is not defined under networks", shared)
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("shared network name must not be empty")
	}
	if _, present := networks[name]; present && shared != name {
		return fmt.Errorf("shared network %q collides with template network %q", shared, name)
	}
	used := false
	for _, entry := range t.Compose.Section("services") {
		service := entry.(map[string]any)
		switch refs := service["networks"].(type) {
		case []any:
			for i, ref := range refs {
				if ref == shared {
					refs[i] = name
					used = true
				}
			}
		case map[string]any:
			if attributes, present := refs[shared]; present {
				if _, collision := refs[name]; collision && shared != name {
					return fmt.Errorf("service network %q collides with shared network", name)
				}
				delete(refs, shared)
				refs[name] = attributes
				used = true
			}
		}
	}
	delete(networks, shared)
	if !used {
		if len(networks) == 0 {
			delete(t.Compose, "networks")
		}
		return nil
	}
	bound := map[string]any{}
	if properties, ok := definition.(map[string]any); ok && !external {
		for key, value := range properties {
			bound[key] = value
		}
	}
	bound["name"] = name
	if external {
		bound["external"] = true
	} else {
		delete(bound, "external")
	}
	networks[name] = bound
	return nil
}
