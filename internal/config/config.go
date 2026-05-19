package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type CLIConfig struct {
	APIKey          string `json:"api_key,omitempty"`
	Domain          string `json:"domain,omitempty"`
	APIURL          string `json:"api_url,omitempty"`
	DefaultTemplate string `json:"default_template,omitempty"`
	DefaultTimeout  int    `json:"default_timeout,omitempty"`
}

func Dir() string {
	if d := os.Getenv("DECLAW_CONFIG_DIR"); d != "" {
		return d
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), ".declaw")
	}
	return filepath.Join(home, ".declaw")
}

func Path() string {
	return filepath.Join(Dir(), "config.json")
}

func Load() (*CLIConfig, error) {
	data, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			return &CLIConfig{}, nil
		}
		return nil, err
	}
	var cfg CLIConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *CLIConfig) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(Path(), data, 0600)
}

type Resolved struct {
	APIKey string
	Domain string
	APIURL string
}

func Resolve(flagAPIKey, flagDomain string) (*Resolved, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	r := &Resolved{
		APIKey: cfg.APIKey,
		Domain: cfg.Domain,
		APIURL: cfg.APIURL,
	}

	if v := os.Getenv("DECLAW_API_KEY"); v != "" {
		r.APIKey = v
	}
	if v := os.Getenv("DECLAW_DOMAIN"); v != "" {
		r.Domain = v
	}
	if v := os.Getenv("DECLAW_API_URL"); v != "" {
		r.APIURL = v
	}

	if flagAPIKey != "" {
		r.APIKey = flagAPIKey
	}
	if flagDomain != "" {
		r.Domain = flagDomain
	}

	return r, nil
}
