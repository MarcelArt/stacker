package models

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
