package models

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Network struct {
	Name     string `yaml:"name"`
	Driver   string `yaml:"driver"`
	External bool   `yaml:"external"`
}

type Volume struct {
	Name     string `yaml:"name"`
	Driver   string `yaml:"driver,omitempty"`
	External bool   `yaml:"external"`
}

type Healthcheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty"`
	Retries  int      `yaml:"retries,omitempty"`
}

type Service struct {
	Image         string            `yaml:"image"`
	ContainerName string            `yaml:"container_name"`
	Networks      []string          `yaml:"networks,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Restart       string            `yaml:"restart"`
	Command       []string          `yaml:"command,omitempty"`
	Entrypoint    []string          `yaml:"entrypoint,omitempty"`
	Expose        []string          `yaml:"expose,omitempty"`
	Healthcheck   *Healthcheck      `yaml:"healthcheck,omitempty"`
	Links         []string          `yaml:"links,omitempty"`
	DependsOn     []string          `yaml:"depends_on,omitempty"`
	User          string            `yaml:"user,omitempty"`
}

type Compose struct {
	Networks map[string]Network `yaml:"networks"`
	Volumes  map[string]Volume  `yaml:"volumes"`
	Services map[string]Service `yaml:"services"`
}

func (c *Compose) SetNetwork(key string, value Network) {
	if c.Networks == nil {
		c.Networks = make(map[string]Network)
	}
	c.Networks[key] = value
}

func (c *Compose) SetVolume(key string, value Volume) {
	if c.Volumes == nil {
		c.Volumes = make(map[string]Volume)
	}
	c.Volumes[key] = value
}

func (c *Compose) SetService(key string, value Service) {
	if c.Services == nil {
		c.Services = make(map[string]Service)
	}
	c.Services[key] = value
}

func (s *Service) UnmarshalYAML(value *yaml.Node) error {
	type Alias struct {
		Image         string            `yaml:"image"`
		ContainerName string            `yaml:"container_name"`
		Networks      []string          `yaml:"networks,omitempty"`
		Volumes       []string          `yaml:"volumes,omitempty"`
		Ports         []string          `yaml:"ports,omitempty"`
		Restart       string            `yaml:"restart"`
		Expose        []string          `yaml:"expose,omitempty"`
		Healthcheck   *Healthcheck      `yaml:"healthcheck,omitempty"`
		Links         []string          `yaml:"links,omitempty"`
		DependsOn     []string          `yaml:"depends_on,omitempty"`
		User          string            `yaml:"user,omitempty"`
	}
	var aux struct {
		Alias       `yaml:",inline"`
		Command     yaml.Node `yaml:"command,omitempty"`
		Entrypoint  yaml.Node `yaml:"entrypoint,omitempty"`
		Environment yaml.Node `yaml:"environment,omitempty"`
	}

	if err := value.Decode(&aux); err != nil {
		return err
	}

	s.Image = aux.Image
	s.ContainerName = aux.ContainerName
	s.Networks = aux.Networks
	s.Volumes = aux.Volumes
	s.Ports = aux.Ports
	s.Restart = aux.Restart
	s.Expose = aux.Expose
	s.Healthcheck = aux.Healthcheck
	s.Links = aux.Links
	s.DependsOn = aux.DependsOn
	s.User = aux.User

	if aux.Command.Kind != 0 {
		var str string
		if err := aux.Command.Decode(&str); err == nil {
			s.Command = []string{str}
		} else {
			var slice []string
			if err := aux.Command.Decode(&slice); err == nil {
				s.Command = slice
			} else {
				return fmt.Errorf("failed to decode command")
			}
		}
	}

	if aux.Entrypoint.Kind != 0 {
		var str string
		if err := aux.Entrypoint.Decode(&str); err == nil {
			s.Entrypoint = []string{str}
		} else {
			var slice []string
			if err := aux.Entrypoint.Decode(&slice); err == nil {
				s.Entrypoint = slice
			} else {
				return fmt.Errorf("failed to decode entrypoint")
			}
		}
	}

	if aux.Environment.Kind != 0 {
		var mp map[string]string
		if err := aux.Environment.Decode(&mp); err == nil {
			s.Environment = mp
		} else {
			var slice []string
			if err := aux.Environment.Decode(&slice); err == nil {
				res := make(map[string]string)
				for _, item := range slice {
					parts := strings.SplitN(item, "=", 2)
					if len(parts) == 2 {
						res[parts[0]] = parts[1]
					} else {
						res[parts[0]] = ""
					}
				}
				s.Environment = res
			} else {
				return fmt.Errorf("failed to decode environment")
			}
		}
	}

	return nil
}
