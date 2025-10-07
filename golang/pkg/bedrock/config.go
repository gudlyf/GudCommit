package bedrock

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds optional runtime configuration loaded from a JSON file or env vars.
type Config struct {
	ModelID        string `json:"model_id"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Region         string `json:"region"`
}

const (
	defaultModelID = "anthropic.claude-3-5-sonnet-20240620-v1:0"
)

// loadConfig attempts to load configuration from environment variables and JSON files.
// Precedence: env vars > ~/.gudcommit.json > ~/.gudchangelog.json > defaults
func loadConfig() (Config, error) {
	cfg := Config{
		ModelID:        defaultModelID,
		TimeoutSeconds: 60,
		Region:         DefaultAWSRegion,
	}

	// Load from files if present
	home, err := os.UserHomeDir()
	if err == nil {
		candidates := []string{
			filepath.Join(home, ".gudcommit.json"),
			filepath.Join(home, ".gudchangelog.json"),
		}
		for _, p := range candidates {
			if _, statErr := os.Stat(p); statErr == nil {
				f, openErr := os.Open(p)
				if openErr != nil {
					return cfg, fmt.Errorf("failed to open config file %s: %w", p, openErr)
				}
				defer f.Close()
				var fileCfg Config
				if decErr := json.NewDecoder(f).Decode(&fileCfg); decErr != nil {
					return cfg, fmt.Errorf("failed to decode config file %s: %w", p, decErr)
				}
				// Merge file config
				if fileCfg.ModelID != "" {
					cfg.ModelID = fileCfg.ModelID
				}
				if fileCfg.TimeoutSeconds > 0 {
					cfg.TimeoutSeconds = fileCfg.TimeoutSeconds
				}
				if fileCfg.Region != "" {
					cfg.Region = fileCfg.Region
				}
				break
			}
		}
	}

	// Env var overrides
	if v := os.Getenv("GUD_BEDROCK_MODEL_ID"); v != "" {
		cfg.ModelID = v
	}
	if v := os.Getenv("GUD_HTTP_TIMEOUT_SECONDS"); v != "" {
		if n, convErr := atoiStrict(v); convErr == nil && n > 0 {
			cfg.TimeoutSeconds = n
		}
	}
	if v := os.Getenv("AWS_REGION"); v != "" {
		cfg.Region = v
	}

	return cfg, nil
}

// atoiStrict converts string to int without accepting leading/trailing spaces.
func atoiStrict(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty string")
	}
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid digit: %q", c)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
