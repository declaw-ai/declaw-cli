package cmdutil

import (
	"fmt"
	"os"
	"strings"

	"github.com/declaw-ai/declaw-cli/internal/config"
	declaw "github.com/declaw-ai/declaw-go"
	"github.com/spf13/cobra"
)

func ResolveConfig(cmd *cobra.Command) (*config.Resolved, error) {
	flagKey, _ := cmd.Flags().GetString("api-key")
	flagDomain, _ := cmd.Flags().GetString("domain")
	return config.Resolve(flagKey, flagDomain)
}

func JSONOutput(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("json")
	return v
}

func SandboxOpts(cfg *config.Resolved) []declaw.SandboxOption {
	var opts []declaw.SandboxOption
	if cfg.APIKey != "" {
		opts = append(opts, declaw.WithSandboxAPIKey(cfg.APIKey))
	}
	if cfg.APIURL != "" {
		opts = append(opts, declaw.WithSandboxAPIURL(cfg.APIURL))
	}
	if cfg.Domain != "" {
		opts = append(opts, declaw.WithSandboxDomain(cfg.Domain))
	}
	return opts
}

func ListOpts(cfg *config.Resolved) []declaw.ListOption {
	var opts []declaw.ListOption
	if cfg.APIKey != "" {
		opts = append(opts, declaw.WithListAPIKey(cfg.APIKey))
	}
	if cfg.APIURL != "" {
		opts = append(opts, declaw.WithListAPIURL(cfg.APIURL))
	}
	if cfg.Domain != "" {
		opts = append(opts, declaw.WithListDomain(cfg.Domain))
	}
	return opts
}

func AccountOpts(cfg *config.Resolved) []declaw.ConfigOption {
	var opts []declaw.ConfigOption
	if cfg.APIKey != "" {
		opts = append(opts, declaw.WithAPIKey(cfg.APIKey))
	}
	if cfg.APIURL != "" {
		opts = append(opts, declaw.WithAPIURL(cfg.APIURL))
	}
	if cfg.Domain != "" {
		opts = append(opts, declaw.WithDomain(cfg.Domain))
	}
	return opts
}

func VaultOpts(cfg *config.Resolved) []declaw.ConfigOption {
	var opts []declaw.ConfigOption
	if cfg.APIKey != "" {
		opts = append(opts, declaw.WithAPIKey(cfg.APIKey))
	}
	if cfg.APIURL != "" {
		opts = append(opts, declaw.WithAPIURL(cfg.APIURL))
	}
	if cfg.Domain != "" {
		opts = append(opts, declaw.WithDomain(cfg.Domain))
	}
	return opts
}

func RestoreOpts(cfg *config.Resolved) []declaw.RestoreOption {
	var opts []declaw.RestoreOption
	if cfg.APIKey != "" {
		opts = append(opts, declaw.WithRestoreAPIKey(cfg.APIKey))
	}
	if cfg.APIURL != "" {
		opts = append(opts, declaw.WithRestoreAPIURL(cfg.APIURL))
	}
	return opts
}

func ParseKeyValues(pairs []string) (map[string]string, error) {
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid KEY=VALUE format: %q", p)
		}
		if k == "" {
			return nil, fmt.Errorf("invalid KEY=VALUE format: %q: key cannot be empty", p)
		}
		m[k] = v
	}
	return m, nil
}

func ParseEnvPairs(pairs []string) (map[string]string, error) {
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if ok {
			if k == "" {
				return nil, fmt.Errorf("invalid env var: %q: key cannot be empty", p)
			}
			m[k] = v
			continue
		}
		v, found := os.LookupEnv(p)
		if !found {
			return nil, fmt.Errorf("environment variable %q not set", p)
		}
		m[p] = v
	}
	return m, nil
}
