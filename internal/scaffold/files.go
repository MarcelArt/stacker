package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func encodeCompose(doc Document) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(doc); err != nil {
		return nil, fmt.Errorf("encode Compose: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func stageFile(path string, data []byte, defaultMode os.FileMode) (string, error) {
	mode := defaultMode
	if info, err := os.Stat(path); err == nil {
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("%s is not a regular file", path)
		}
		mode = info.Mode().Perm()
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".stacker-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	success := false
	defer func() {
		file.Close()
		if !success {
			os.Remove(name)
		}
	}()
	if err := file.Chmod(mode); err != nil {
		return "", err
	}
	if _, err := file.Write(data); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	success = true
	return name, nil
}

func WriteCompose(path string, doc Document) error {
	data, err := encodeCompose(doc)
	if err != nil {
		return err
	}
	staged, err := stageFile(path, data, 0644)
	if err != nil {
		return err
	}
	defer os.Remove(staged)
	return os.Rename(staged, path)
}

func EnvironmentPath(composePath string) string {
	return filepath.Join(filepath.Dir(composePath), ".env")
}

// WriteStack stages both files before publishing either. If publishing Compose
// fails, restore the previous environment file to avoid a partially applied add.
func WriteStack(path string, doc Document, environment []byte) error {
	envPath := EnvironmentPath(path)
	if filepath.Clean(path) == envPath {
		return fmt.Errorf("Compose file and environment file must have different paths")
	}
	data, err := encodeCompose(doc)
	if err != nil {
		return err
	}
	composeStage, err := stageFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("stage Compose file: %w", err)
	}
	defer os.Remove(composeStage)
	oldEnv, readErr := os.ReadFile(envPath)
	if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
		return readErr
	}
	if readErr == nil && bytes.Equal(oldEnv, environment) {
		return os.Rename(composeStage, path)
	}
	envStage, err := stageFile(envPath, environment, 0600)
	if err != nil {
		return fmt.Errorf("stage environment file: %w", err)
	}
	defer os.Remove(envStage)
	if err := os.Rename(envStage, envPath); err != nil {
		return fmt.Errorf("write environment file: %w", err)
	}
	if err := os.Rename(composeStage, path); err != nil {
		var rollbackErr error
		if errors.Is(readErr, os.ErrNotExist) {
			rollbackErr = os.Remove(envPath)
		} else {
			var backup string
			backup, rollbackErr = stageFile(envPath, oldEnv, 0600)
			if rollbackErr == nil {
				rollbackErr = os.Rename(backup, envPath)
				os.Remove(backup)
			}
		}
		return errors.Join(fmt.Errorf("write Compose file: %w", err), rollbackErr)
	}
	return nil
}
