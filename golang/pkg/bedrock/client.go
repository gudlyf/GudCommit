package bedrock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	DefaultAWSRegion = "us-east-1"
)

// BedrockRequest represents the request payload for Bedrock
type BedrockRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []Message `json:"messages"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// BedrockResponse represents the response from Bedrock
type BedrockResponse struct {
	Content []ContentBlock `json:"content"`
	Usage   Usage          `json:"usage"`
}

// ContentBlock represents a content block in the response
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Usage represents token usage information
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Client represents a Bedrock API client
type Client struct {
	APIKey string
	Region string
}

// NewClient creates a new Bedrock client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("GUD_BEDROCK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GUD_BEDROCK_API_KEY environment variable is not set")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = DefaultAWSRegion
	}

	return &Client{
		APIKey: apiKey,
		Region: region,
	}, nil
}

// InvokeModel invokes Bedrock directly using API key authentication
func (c *Client) InvokeModel(prompt, repoPath string) (string, error) {
	// Load dynamic configuration (model ID, timeout, region)
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}

	// Construct the Bedrock endpoint for Claude
	endpoint := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke", cfg.Region, cfg.ModelID)

	// Create the request payload for Claude
	payload := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Marshal to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for API key authentication
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	// Create a simple spinner that updates in place with timer
	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan bool)
	startTime := time.Now()

	// Start spinner animation in a goroutine
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				elapsed := time.Since(startTime)
				seconds := int(elapsed.Seconds())
				fmt.Printf("\r\033[K%s :: Awaiting response from Bedrock ... [%ds]", spinnerChars[i%len(spinnerChars)], seconds)
				os.Stdout.Sync()
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Make the request
	client := &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)

	// Signal completion and clear the spinner line
	done <- true
	fmt.Print("\r\033[K") // Clear the spinner line

	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bedrock API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response BedrockResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the content from Claude's response
	if len(response.Content) > 0 {
		fmt.Print("✔ :: Response received from Bedrock\n")
		return response.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in response")
}
