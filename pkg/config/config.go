package config

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v3"
)

const configFileName = "config.yaml"

type Config struct {
	ServerID     string `yaml:"server-id"`
	Token        string `yaml:"token"`
	ClientID     string `yaml:"client-id"`
	Permissions  string `yaml:"permissions"`
	ClientSecret string `yaml:"client-secret"`
	RedirectURL  string `yaml:"redirect-url"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = configFileName
	}

	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %w", configFileName, err)
	}

	var cfg Config
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal config file: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("token field cannot be empty")
	}

	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client-id field cannot be empty")
	}

	if cfg.Permissions == "" {
		return nil, fmt.Errorf("permissions field cannot be empty")
	}

	if cfg.RedirectURL == "" {
		return nil, fmt.Errorf("redirect-url field cannot be empty")
	}

	return &cfg, nil
}
