package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ComposeFile represents a parsed docker-compose.yml file
type ComposeFile struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services,omitempty"`
	Networks map[string]Network `yaml:"networks,omitempty"`
	Volumes  map[string]Volume  `yaml:"volumes,omitempty"`
}

// Service represents a service in docker-compose.yml
type Service struct {
	Name        string            `yaml:"-"` // Don't marshal, used as map key
	Image       string            `yaml:"image,omitempty"`
	Build       *BuildConfig      `yaml:"build,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Volumes     []string          `yaml:"volumes,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Networks    []string          `yaml:"networks,omitempty"`
	HealthCheck *HealthCheck      `yaml:"healthcheck,omitempty"`
	Restart     string            `yaml:"restart,omitempty"`
	Command     interface{}       `yaml:"command,omitempty"` // string or []string
	Entrypoint  interface{}       `yaml:"entrypoint,omitempty"`
	WorkingDir  string            `yaml:"working_dir,omitempty"`
	User        string            `yaml:"user,omitempty"`
	Hostname    string            `yaml:"hostname,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

// BuildConfig represents build configuration for a service
type BuildConfig struct {
	Context    string            `yaml:"context,omitempty"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Args       map[string]string `yaml:"args,omitempty"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Test     []string `yaml:"test,omitempty"`
	Interval string   `yaml:"interval,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty"`
	Retries  int      `yaml:"retries,omitempty"`
}

// Network represents a network in docker-compose.yml
type Network struct {
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty"`
}

// Volume represents a volume in docker-compose.yml
type Volume struct {
	Driver     string            `yaml:"driver,omitempty"`
	DriverOpts map[string]string `yaml:"driver_opts,omitempty"`
	External   bool              `yaml:"external,omitempty"`
}

// ParseComposeFile parses a docker-compose.yml file
func ParseComposeFile(path string) (*ComposeFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &compose, nil
}

// WriteComposeFile writes the compose file to disk
func (c *ComposeFile) WriteComposeFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ServiceExists checks if a service with the given name exists
func (c *ComposeFile) ServiceExists(name string) bool {
	_, exists := c.Services[name]
	return exists
}

// NetworkExists checks if a network with the given name exists
func (c *ComposeFile) NetworkExists(name string) bool {
	_, exists := c.Networks[name]
	return exists
}

// VolumeExists checks if a volume with the given name exists
func (c *ComposeFile) VolumeExists(name string) bool {
	_, exists := c.Volumes[name]
	return exists
}

// AddService adds a service to the compose file
func (c *ComposeFile) AddService(service Service) {
	if c.Services == nil {
		c.Services = make(map[string]Service)
	}
	c.Services[service.Name] = service
}

// AddNetwork adds a network to the compose file
func (c *ComposeFile) AddNetwork(name string, network Network) {
	if c.Networks == nil {
		c.Networks = make(map[string]Network)
	}
	c.Networks[name] = network
}

// AddVolume adds a volume to the compose file
func (c *ComposeFile) AddVolume(name string, volume Volume) {
	if c.Volumes == nil {
		c.Volumes = make(map[string]Volume)
	}
	c.Volumes[name] = volume
}

// Validate validates the compose file structure
func (cf *ComposeFile) Validate() error {
	// TODO: Implement validation
	return nil
}

// GetDependencyGraph builds a dependency graph of services
func (cf *ComposeFile) GetDependencyGraph() (map[string][]string, error) {
	// TODO: Implement dependency graph generation
	return nil, nil
}