package config

import (
	"os"
	"path/filepath"
)

const (
	// ConfigFileName is the name of the project configuration file
	ConfigFileName = ".container-composer.yaml"
)

// Config represents the project configuration
type Config struct {
	ProjectName  string            `yaml:"project_name"`
	ComposeFiles []string          `yaml:"compose_files"`
	Environment  string            `yaml:"environment"`
	Environments map[string]EnvConfig `yaml:"environments"`
	Plugins      []string          `yaml:"plugins"`
}

// EnvConfig represents environment-specific configuration
type EnvConfig struct {
	ComposeFiles []string          `yaml:"compose_files"`
	EnvFiles     []string          `yaml:"env_files"`
	Variables    map[string]string `yaml:"variables"`
}

// Load loads the configuration from the current directory
func Load() (*Config, error) {
	// TODO: Implement config loading
	return &Config{
		ProjectName: "my-project",
		ComposeFiles: []string{"docker-compose.yml"},
		Environment: "dev",
	}, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	// TODO: Implement config saving
	return nil
}

// FindConfigFile searches for the config file in current and parent directories
func FindConfigFile() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Search up the directory tree
	dir := cwd
	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}