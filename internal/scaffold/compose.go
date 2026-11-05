package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Document preserves Compose attributes without restricting them to a Go schema.
type Document map[string]any

func ParseCompose(data []byte) (Document, error) {
	// Decode into the unnamed map type so yaml.v3 uses map[string]any for
	// nested mappings as well, rather than propagating the Document type.
	raw := map[string]any{}
	if len(bytes.TrimSpace(data)) == 0 {
		return Document{}, nil
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("invalid Compose YAML: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("expected one YAML document")
	}
	doc := Document(raw)
	if doc == nil {
		return nil, fmt.Errorf("Compose must be a mapping")
	}
	for _, section := range []string{"services", "networks", "volumes", "configs", "secrets"} {
		value, present := doc[section]
		if !present {
			continue
		}
		entries, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s must be a mapping", section)
		}
		for name, entry := range entries {
			if _, ok := entry.(map[string]any); !ok && (entry != nil || section == "services") {
				return nil, fmt.Errorf("%s.%s must be a mapping", section, name)
			}
		}
	}
	return doc, nil
}

func ReadCompose(path string, allowMissing bool) (Document, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) && allowMissing {
		return Document{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read Compose file: %w", err)
	}
	return ParseCompose(data)
}

func (d Document) Section(name string) map[string]any {
	section, _ := d[name].(map[string]any)
	return section
}

// Merge returns a new document, leaving both input documents untouched on failure.
func Merge(existing, incoming Document, overwrite bool) (Document, error) {
	result := Document{}
	for key, value := range existing {
		result[key] = value
	}
	for _, section := range sortedKeys(incoming) {
		value := incoming[section]
		switch section {
		case "services", "networks", "volumes", "configs", "secrets":
			entries := map[string]any{}
			for key, entry := range existing.Section(section) {
				entries[key] = entry
			}
			for _, name := range sortedKeys(incoming.Section(section)) {
				entry := incoming.Section(section)[name]
				if old, present := entries[name]; present {
					if section == "services" {
						if !overwrite {
							return nil, fmt.Errorf("service %q already exists; use --overwrite to replace it", name)
						}
					} else {
						if !reflect.DeepEqual(resourceValue(old), resourceValue(entry)) {
							return nil, fmt.Errorf("conflicting %s definition for %q", section, name)
						}
						continue
					}
				}
				entries[name] = entry
			}
			result[section] = entries
		default:
			if old, present := existing[section]; present && !reflect.DeepEqual(old, value) {
				return nil, fmt.Errorf("conflicting top-level Compose field %q", section)
			}
			result[section] = value
		}
	}
	if err := validateNetworking(result); err != nil {
		return nil, err
	}
	return result, nil
}

// Treat omitted false/default resource properties and empty resource definitions equally.
func resourceValue(value any) map[string]any {
	result := map[string]any{}
	if fields, ok := value.(map[string]any); ok {
		for key, field := range fields {
			if key == "external" && field == false {
				continue
			}
			result[key] = field
		}
	}
	return result
}

func validateNetworking(doc Document) error {
	for _, name := range sortedKeys(doc.Section("services")) {
		service := doc.Section("services")[name].(map[string]any)
		mode, hasMode := service["network_mode"]
		if hasMode {
			if text, ok := mode.(string); !ok || text == "" {
				return fmt.Errorf("service %q network_mode must be a nonempty string", name)
			}
			if _, present := service["networks"]; present {
				return fmt.Errorf("service %q cannot combine network_mode and networks", name)
			}
			if mode == "host" {
				if ports, present := service["ports"]; present && !emptyValue(ports) {
					return fmt.Errorf("service %q cannot publish ports with network_mode: host", name)
				}
			}
		}
		if refs, present := service["networks"]; present {
			switch networks := refs.(type) {
			case []any:
				for _, ref := range networks {
					if _, ok := ref.(string); !ok {
						return fmt.Errorf("service %q networks must contain network names", name)
					}
				}
			case map[string]any:
			default:
				return fmt.Errorf("service %q networks must be a list or mapping", name)
			}
		}
	}
	return nil
}

func emptyValue(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	return (v.Kind() == reflect.Slice || v.Kind() == reflect.Map || v.Kind() == reflect.String) && v.Len() == 0
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// MissingDependencies reports advisory dependencies without preventing a partial stack.
func MissingDependencies(doc Document, added Document) []string {
	missing := map[string]bool{}
	for _, value := range added.Section("services") {
		service := value.(map[string]any)
		dependencies := []string{}
		switch refs := service["depends_on"].(type) {
		case []any:
			for _, ref := range refs {
				if name, ok := ref.(string); ok {
					dependencies = append(dependencies, name)
				}
			}
		case map[string]any:
			dependencies = append(dependencies, sortedKeys(refs)...)
		}
		for _, dependency := range dependencies {
			if _, present := doc.Section("services")[dependency]; !present {
				missing[dependency] = true
			}
		}
	}
	return sortedKeys(missing)
}

// PortSummary supports both Compose short and long port syntax.
func PortSummary(service map[string]any) string {
	var ports []string
	if entries, ok := service["ports"].([]any); ok {
		for _, entry := range entries {
			if port, ok := entry.(map[string]any); ok {
				text := fmt.Sprint(port["target"])
				if published, ok := port["published"]; ok {
					text = fmt.Sprint(published) + ":" + text
				}
				if host, ok := port["host_ip"]; ok {
					text = fmt.Sprint(host) + ":" + text
				}
				if protocol, ok := port["protocol"]; ok {
					text += "/" + fmt.Sprint(protocol)
				}
				ports = append(ports, text)
			} else {
				ports = append(ports, fmt.Sprint(entry))
			}
		}
	}
	if len(ports) == 0 {
		return "-"
	}
	return strings.Join(ports, ", ")
}
