package bedrock

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	// Bedrock serves current Claude models through cross-region inference
	// profiles; the "us." profile matches the default us-east-1 region.
	defaultModelID        = "us.anthropic.claude-opus-5-5"
	defaultAnthropicModel = "claude-opus-5-5"
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
				debugf("Loading config from: %s\n", p)

				data, readErr := os.ReadFile(p)
				if readErr != nil {
					return cfg, fmt.Errorf("failed to read config file %s: %w", p, readErr)
				}

				var fileCfg Config
				if decErr := json.Unmarshal(data, &fileCfg); decErr != nil {
					// Do not wrap decErr: json syntax errors can quote file contents (secrets).
					return cfg, fmt.Errorf("failed to decode config file %s: invalid JSON", p)
				}

				if fileCfg.AnthropicAPIKey != "" {
					warnIfInsecurePerms(p)
				}

				// Merge file config
				if fileCfg.Provider != "" {
					cfg.Provider = normalizeProvider(fileCfg.Provider)
				}
				if fileCfg.ModelID != "" {
					cfg.ModelID = fileCfg.ModelID
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
		debugf("ModelID overridden by GUD_BEDROCK_MODEL_ID env var: %s\n", cfg.ModelID)
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

	// Never log secrets here (AnthropicAPIKey).
	debugf("Final configuration: provider=%s model=%s timeout=%ds region=%s\n",
		cfg.Provider, cfg.ModelID, cfg.TimeoutSeconds, cfg.Region)

	return cfg, nil
}

// debugEnabled reports whether GUD_DEBUG is set to a truthy value.
func debugEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GUD_DEBUG"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

// debugf writes diagnostic output to stderr only when GUD_DEBUG is enabled.
func debugf(format string, args ...any) {
	if debugEnabled() {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// warnIfInsecurePerms warns when a file holding secrets is readable by group or others.
func warnIfInsecurePerms(path string) {
	// Unix permission bits are not meaningful on Windows (files report 0666).
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		fmt.Fprintf(os.Stderr, "⚠️  %s contains an API key but has permissions %#o; run: chmod 600 %s\n", path, mode, path)
	}
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
