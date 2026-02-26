package bedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Anthropic Messages API request/response types (subset we need)
// https://docs.anthropic.com/en/api/messages

type anthropicMessageReq struct {
	Model      string         `json:"model"`
	MaxTokens  int            `json:"max_tokens"`
	Messages   []anthropicMsg `json:"messages"`
	APIVersion string         `json:"anthropic_version,omitempty"`
}

type anthropicMsg struct {
	Role    string           `json:"role"`
	Content []anthropicBlock `json:"content"`
}

type anthropicBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type anthropicMessageResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	} `json:"content"`
	StopReason string `json:"stop_reason,omitempty"`
}

const anthropicAPIBase = "https://api.anthropic.com/v1/messages"
const anthropicAPIVersion = "2023-06-01"

// AnthropicClient calls the Anthropic Messages API (same backend as Claude app / Claude Code).
type AnthropicClient struct {
	apiKey     string
	modelID    string
	timeoutSec int
	httpClient *http.Client
}

// NewAnthropicClient creates a client for the Anthropic API using config and ANTHROPIC_API_KEY.
func NewAnthropicClient(cfg Config) (*AnthropicClient, error) {
	apiKey := cfg.AnthropicAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("GUD_ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf(`Anthropic provider requires an API key.

Set one of:
  - ANTHROPIC_API_KEY (or GUD_ANTHROPIC_API_KEY)
  - Or "anthropic_api_key" in ~/.gudcommit.json

Get an API key from: https://console.anthropic.com/
Claude Pro/Max/Teams/Enterprise subscribers have API access.`)
	}

	modelID := cfg.ModelID
	if modelID == "" {
		modelID = defaultAnthropicModel
	}

	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 60
	}

	return &AnthropicClient{
		apiKey:     apiKey,
		modelID:    modelID,
		timeoutSec: timeout,
		httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}, nil
}

// InvokeModel sends the prompt to the Anthropic Messages API and returns the first text block.
func (c *AnthropicClient) InvokeModel(prompt, _ string) (string, error) {
	payload := anthropicMessageReq{
		Model:      c.modelID,
		MaxTokens:  2048,
		APIVersion: anthropicAPIVersion,
		Messages: []anthropicMsg{
			{
				Role: "user",
				Content: []anthropicBlock{
					{Type: "text", Text: prompt},
				},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("anthropic request marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		anthropicAPIBase,
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("anthropic request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	// Spinner
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan bool)
	start := time.Now()
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(start)
				fmt.Fprintf(os.Stderr, "\r\033[K%s :: Awaiting response from Claude (Anthropic API) ... [%ds]", spinnerChars[i%len(spinnerChars)], int(elapsed.Seconds()))
				os.Stderr.Sync()
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	resp, err := c.httpClient.Do(req)
	done <- true
	fmt.Fprint(os.Stderr, "\r\033[K")
	if err != nil {
		return "", fmt.Errorf("anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic API error: %d - %s", resp.StatusCode, string(respBody))
	}

	var out anthropicMessageResp
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("anthropic response decode: %w", err)
	}
	for _, block := range out.Content {
		if block.Type == "text" && block.Text != "" {
			fmt.Fprint(os.Stderr, "✔ :: Response received from Claude (Anthropic API)\n")
			return block.Text, nil
		}
	}
	return "", fmt.Errorf("anthropic API returned no text content")
}
