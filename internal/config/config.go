package config

import (
	"errors"

	// Errors package for handling errors.
	"fmt"
	// Fmt package for formatting strings.
	"os"
	// Os package for interacting with the operating system, like file handling.
	"path/filepath"
	// Filepath package for manipulating filename paths.
	"gopkg.in/yaml.v3"
	// Yaml.v3 package for YAML processing.
)

// Config represents the entire configuration as structured in YAML.

type Config struct {
	Environment Environment   `yaml:"environment"`
	Logging     LoggingConfig `yaml:"logging"`
	ProxyConfig ProxyConfig   `yaml:"proxyConfig"`
}
type ProxyConfig struct {
	HTTPPort  string            `yaml:"httpPort"`
	HTTPSPort string            `yaml:"httpsPort"`
	CertFile  string            `yaml:"certFile"`
	KeyFile   string            `yaml:"keyFile"`
	Redirects map[string]string `yaml:"redirects"`
}
type LoggingConfig struct {
	Level      LogLevel   `yaml:"level"`
	FileOutput FileConfig `yaml:"fileOutput"`
}

type FileConfig struct {
	FilePath       string         `yaml:"filePath"`
	RotationPolicy RotationPolicy `yaml:"rotationPolicy"`
	MaxSizeMB      int            `yaml:"maxSizeMB"`
	MaxBackups     int            `yaml:"maxBackups"`
}

// LoadConfig reads and decodes the YAML configuration file based on the APP_ENV environment variable.
func LoadConfig(configPath string, configName string) (*Config, error) {
	configFile := filepath.Join(configPath, configName)

	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("config file does not exist: %s", configFile)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}
	// Expand environment variables in the YAML
	expandedData := []byte(os.ExpandEnv(string(data)))
	// Declares a variable of type T to hold the configuration data.

	var config Config
	if err := yaml.Unmarshal(expandedData, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &config, nil
}
