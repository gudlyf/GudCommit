package bedrock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/smithy-go"
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
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      Usage          `json:"usage"`
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
	bedrockClient *bedrockruntime.Client
	Region        string
	Profile       string
}

// NewClient creates a new Bedrock client using AWS SDK with login credentials support
func NewClient() (*Client, error) {
	ctx := context.Background()

	// Determine region
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = DefaultAWSRegion
	}

	// Determine profile to use
	// Default to "gudcommit" profile if AWS_PROFILE is not set
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "gudcommit"
	}

	// Load AWS configuration using the standard credential chain: environment
	// variables, static keys, `aws login` (login_session), IAM Identity Center
	// (SSO), assumed roles, credential_process, and container/instance roles.
	// See golang/QUICKSTART_AWS_LOGIN.md.
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithSharedConfigProfile(profile),
		// Enable retry behavior for transient errors
		config.WithRetryMaxAttempts(3),
	)
	if err != nil {
		return nil, enhanceConfigError(err, profile, region)
	}

	// Create Bedrock Runtime client
	bedrockClient := bedrockruntime.NewFromConfig(cfg)

	return &Client{
		bedrockClient: bedrockClient,
		Region:        region,
		Profile:       profile,
	}, nil
}

// enhanceConfigError explains how to set up AWS credentials when config loading fails.
func enhanceConfigError(err error, profile, region string) error {
	return fmt.Errorf(`unable to load AWS configuration (profile %q, region %q)

Set up the profile with aws login (AWS CLI v2.32.0+):
  aws login --profile %s --region %s

Or use another profile:  export AWS_PROFILE=<profile>
See golang/QUICKSTART_AWS_LOGIN.md for other credential options.

Original error: %w`, profile, region, profile, region, err)
}

// Invoker is the common interface for any LLM provider (Bedrock or Anthropic).
type Invoker interface {
	InvokeModel(prompt, repoPath string) (string, error)
}

// NewInvoker loads config and returns the configured provider (Bedrock, Anthropic API, or Claude CLI).
func NewInvoker() (Invoker, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	var inv Invoker
	switch cfg.Provider {
	case ProviderClaudeCLI:
		inv, err = NewClaudeCLIClient(cfg)
	case ProviderAnthropic:
		inv, err = NewAnthropicClient(cfg)
	default:
		inv, err = NewClient()
	}
	if err != nil {
		return nil, err
	}
	return sanitizingInvoker{inner: inv}, nil
}

// InvokeModel invokes Bedrock using AWS SDK with SSO credential support
func (c *Client) InvokeModel(prompt, repoPath string) (string, error) {
	// Load dynamic configuration (model ID, timeout, region)
	cfg, err := LoadConfig()
	if err != nil {
		return "", err
	}

	// Create the request payload for Claude
	payload := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        defaultMaxTokens,
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	// Invoke the model using AWS SDK
	// This automatically handles SigV4 signing with SSO credentials
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(cfg.ModelID),
		Body:        jsonPayload,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}

	result, err := c.bedrockClient.InvokeModel(ctx, input)

	// Signal completion and clear the spinner line
	done <- true
	fmt.Print("\r\033[K") // Clear the spinner line

	if err != nil {
		// Check if it's a credential error and provide helpful message
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "UnrecognizedClientException" ||
				apiErr.ErrorCode() == "InvalidSignatureException" {
				return "", enhanceCredentialError(err, c.Profile)
			}
		}

		// Check for other credential-related errors
		errStr := err.Error()
		if containsAny(errStr, []string{
			"no EC2 IMDS role found",
			"failed to refresh cached credentials",
			"get credentials",
			"no valid providers",
			"credential_process",
		}) {
			return "", enhanceCredentialError(err, c.Profile)
		}

		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	// Parse response
	var response BedrockResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	text, err := extractText("Bedrock", response.StopReason, response.Content)
	if err != nil {
		return "", err
	}
	fmt.Print("✔ :: Response received from Bedrock\n")
	return text, nil
}

// enhanceCredentialError explains how to fix missing or expired AWS credentials.
func enhanceCredentialError(err error, profile string) error {
	return fmt.Errorf(`AWS credential error (profile %q): credentials are missing or expired

If you use aws login, sign in again:
  aws login --profile %s

Check which identity is used:  aws sts get-caller-identity --profile %s
See golang/QUICKSTART_AWS_LOGIN.md for other credential options.

Original error: %w`, profile, profile, profile, err)
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(substr) > 0 && len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
