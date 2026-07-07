package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const ConfigFileName = "strata.json"

type ProjectConfig struct {
	Name       string `json:"name"`
	GatewayURL string `json:"gateway_url"`
	ProjectID  string `json:"project_id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type CLIConfig struct {
	CurrentProject string                   `json:"current_project"`
	Projects       map[string]*ProjectConfig `json:"projects"`
}

func LoadCLIConfig() (*CLIConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find home directory: %w", err)
	}
	dir := filepath.Join(home, ".strata")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "config.json")

	cfg := &CLIConfig{Projects: make(map[string]*ProjectConfig)}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	json.Unmarshal(data, cfg)
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]*ProjectConfig)
	}
	return cfg, nil
}

func SaveCLIConfig(cfg *CLIConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".strata")
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func FindProjectConfig() (*ProjectConfig, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		path := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			var cfg ProjectConfig
			if err := json.Unmarshal(data, &cfg); err != nil {
				return nil, err
			}
			return &cfg, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("no %s found in any parent directory", ConfigFileName)
		}
		dir = parent
	}
}

func GetGatewayURL() string {
	cfg, err := FindProjectConfig()
	if err == nil && cfg.GatewayURL != "" {
		return cfg.GatewayURL
	}
	if url := viper.GetString("gateway"); url != "" {
		return url
	}
	return "http://localhost:8000"
}

func GetAccessToken() string {
	cfg, err := FindProjectConfig()
	if err == nil && cfg.AccessToken != "" {
		return cfg.AccessToken
	}
	cliCfg, err := LoadCLIConfig()
	if err == nil && cliCfg.CurrentProject != "" {
		if p, ok := cliCfg.Projects[cliCfg.CurrentProject]; ok {
			return p.AccessToken
		}
	}
	return ""
}
