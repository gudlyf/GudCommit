package bedrock

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Provider is the LLM backend.
const (
	ProviderBedrock   = "bedrock"    // AWS Bedrock
	ProviderAnthropic = "anthropic"  // Anthropic API (console.anthropic.com)
	ProviderClaudeCLI = "claude-cli" // Claude Code CLI (claude -p); uses your Claude app/Code login
)

// Config holds optional runtime configuration loaded from a JSON file or env vars.
type Config struct {
	Provider        string `json:"provider"` // "bedrock" or "anthropic"
	ModelID         string `json:"model_id"` // Bedrock model ID or Anthropic model name
	TimeoutSeconds  int    `json:"timeout_seconds"`
	Region          string `json:"region"`            // AWS region (Bedrock only)
	AnthropicAPIKey string `json:"anthropic_api_key"` // Optional; else ANTHROPIC_API_KEY env
}

const (
	defaultModelID        = "anthropic.claude-3-5-sonnet-20240620-v1:0"
	defaultAnthropicModel = "claude-sonnet-4-20250514"
)

// LoadConfig loads configuration from environment variables and JSON files.
// Precedence: env vars > ~/.gudcommit.json > ~/.gudchangelog.json > defaults
// It is safe to call from multiple places; use the returned Config to create the appropriate provider.
func LoadConfig() (Config, error) {
	cfg := Config{
		Provider:       ProviderBedrock,
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
				// Debug: Print which config file is being loaded
				fmt.Fprintf(os.Stderr, "Loading config from: %s\n", p)

				data, readErr := os.ReadFile(p)
				if readErr != nil {
					return cfg, fmt.Errorf("failed to read config file %s: %w", p, readErr)
				}

				// Debug: Print raw file contents
				fmt.Fprintf(os.Stderr, "Config file contents: %s\n", string(data))

				var fileCfg Config
				if decErr := json.Unmarshal(data, &fileCfg); decErr != nil {
					return cfg, fmt.Errorf("failed to decode config file %s: %w", p, decErr)
				}

				// Debug: Print parsed config
				fmt.Fprintf(os.Stderr, "Parsed config - ModelID: %s, TimeoutSeconds: %d, Region: %s\n",
					fileCfg.ModelID, fileCfg.TimeoutSeconds, fileCfg.Region)

				// Merge file config
				if fileCfg.Provider != "" {
					cfg.Provider = normalizeProvider(fileCfg.Provider)
				}
				if fileCfg.ModelID != "" {
					cfg.ModelID = fileCfg.ModelID
					fmt.Fprintf(os.Stderr, "Using ModelID from config: %s\n", cfg.ModelID)
				}
				if fileCfg.TimeoutSeconds > 0 {
					cfg.TimeoutSeconds = fileCfg.TimeoutSeconds
				}
				if fileCfg.Region != "" {
					cfg.Region = fileCfg.Region
				}
				if fileCfg.AnthropicAPIKey != "" {
					cfg.AnthropicAPIKey = fileCfg.AnthropicAPIKey
				}
				break
			}
		}
	}

	// Env var overrides
	if v := os.Getenv("GUD_PROVIDER"); v != "" {
		cfg.Provider = normalizeProvider(v)
	}
	if v := os.Getenv("GUD_BEDROCK_MODEL_ID"); v != "" {
		cfg.ModelID = v
		fmt.Fprintf(os.Stderr, "ModelID overridden by GUD_BEDROCK_MODEL_ID env var: %s\n", cfg.ModelID)
	}
	if v := os.Getenv("GUD_ANTHROPIC_MODEL"); v != "" {
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
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.AnthropicAPIKey = v
	}
	if v := os.Getenv("GUD_ANTHROPIC_API_KEY"); v != "" {
		cfg.AnthropicAPIKey = v
	}

	// Default model for Anthropic when not set
	if cfg.Provider == ProviderAnthropic && cfg.ModelID == defaultModelID {
		cfg.ModelID = defaultAnthropicModel
	}

	// Debug: Print final configuration
	fmt.Fprintf(os.Stderr, "\n✓ Final configuration:\n")
	fmt.Fprintf(os.Stderr, "  Provider: %s\n", cfg.Provider)
	fmt.Fprintf(os.Stderr, "  Model ID: %s\n", cfg.ModelID)
	fmt.Fprintf(os.Stderr, "  Timeout: %d seconds\n", cfg.TimeoutSeconds)
	fmt.Fprintf(os.Stderr, "  Region: %s\n\n", cfg.Region)

	return cfg, nil
}

// normalizeProvider returns a known provider constant; invalid values default to "bedrock".
func normalizeProvider(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case ProviderClaudeCLI, "claude_cli", "claudecli", "claude":
		return ProviderClaudeCLI
	case ProviderAnthropic, "anthropic-api", "api":
		return ProviderAnthropic
	default:
		return ProviderBedrock
	}
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
