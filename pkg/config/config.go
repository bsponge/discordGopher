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
	ClientSecret string `yaml:"client-secret"`
	OAuthURL     string `yaml:"oauth-url"`
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

	return &cfg, nil
}
