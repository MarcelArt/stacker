package scaffold

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var assignment = regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$`)
var unsetEntry = regexp.MustCompile(`^\s*(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)\s*(?:#.*)?$`)

// PrepareEnvironment preserves existing bytes and appends only undeclared entries.
// Empty names are returned for setup guidance; secret values are never returned.
func PrepareEnvironment(path string, requirements map[string]EnvironmentRequirement) ([]byte, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("read environment file: %w", err)
	}
	empty, err := environmentEntries(string(data))
	if err != nil {
		return nil, nil, err
	}
	var output strings.Builder
	output.Write(data)
	var pending []string
	for _, name := range sortedKeys(requirements) {
		isEmpty, present := empty[name]
		if !present {
			if output.Len() > 0 && !strings.HasSuffix(output.String(), "\n") {
				output.WriteByte('\n')
			}
			for _, line := range strings.Split(requirements[name].Description, "\n") {
				output.WriteString("# " + strings.TrimSuffix(line, "\r") + "\n")
			}
			output.WriteString(name + "=\n")
			isEmpty = true
		}
		if isEmpty {
			pending = append(pending, name)
		}
	}
	return []byte(output.String()), pending, nil
}

func environmentEntries(data string) (map[string]bool, error) {
	entries := map[string]bool{}
	var quote byte
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if quote != 0 {
			quote = remainingQuote(line, quote)
			continue
		}
		match := assignment.FindStringSubmatch(line)
		if match == nil {
			if unset := unsetEntry.FindStringSubmatch(line); unset != nil {
				entries[unset[1]] = true
			}
			continue
		}
		value := strings.TrimSpace(match[2])
		isEmpty := value == "" || strings.HasPrefix(value, "#")
		if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
			quote = remainingQuote(value[1:], value[0])
			if len(value) >= 2 && value[1] == value[0] {
				isEmpty = true
			}
		}
		entries[match[1]] = isEmpty
	}
	if quote != 0 {
		return nil, fmt.Errorf("environment file has an unterminated quoted value")
	}
	return entries, nil
}

func remainingQuote(text string, quote byte) byte {
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' && i+1 < len(text) && (text[i+1] == quote || quote == '"') {
			i++
			continue
		}
		if text[i] == quote {
			return 0
		}
	}
	return quote
}
