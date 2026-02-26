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
	bedrockClient *bedrockruntime.Client
	Region        string
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

	// Load AWS configuration
	// This automatically handles the AWS credential chain:
	// - Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN)
	// - credential_process (supports new 'aws login' via CLI delegation)
	// - Shared credentials file (~/.aws/credentials)
	// - AWS SSO credentials (legacy aws sso login)
	// - IAM roles
	// - EC2 instance metadata
	//
	// For new 'aws login' support:
	// 1. Run: aws login --profile your-login
	// 2. Configure credential_process in ~/.aws/config
	// 3. SDK v2 will use credentials via credential_process workaround
	//
	// Note: Go SDK v2 doesn't natively support login_session yet, but works
	// via credential_process delegation to AWS CLI
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
	}, nil
}

// enhanceConfigError provides helpful setup instructions when AWS config fails
func enhanceConfigError(err error, profile, region string) error {
	return fmt.Errorf(`unable to load AWS credentials

Current profile: %s
Current region: %s

AWS credentials are not properly configured. Please set up credentials using one of these methods:

Method 1: NEW aws login (Recommended)
----------------------------------------
1. Ensure AWS CLI v2.32.0 or later:
   aws --version

2. Log in with your AWS Console credentials:
   aws login --profile my-aws

3. Configure ~/.aws/config with credential_process:

   [profile my-aws]
   login_session = arn:aws:iam::123456789012:user/yourname
   region = %s

   [profile gudcommit]
   credential_process = aws configure export-credentials --profile my-aws --format process
   region = %s

4. Test it:
   export AWS_PROFILE=gudcommit  # or let gudcommit use it by default
   aws sts get-caller-identity

5. Run gudcommit:
   gudcommit

Method 2: Environment Variables
---------------------------------
export AWS_ACCESS_KEY_ID=your-key-id
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=%s
gudcommit

Method 3: Shared Credentials (~/.aws/credentials)
---------------------------------------------------
[default]
aws_access_key_id = your-key-id
aws_secret_access_key = your-secret-key

For more help, see: golang/QUICKSTART_AWS_LOGIN.md

Original error: %v`, profile, region, region, region, region, err)
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
	switch cfg.Provider {
	case ProviderClaudeCLI:
		return NewClaudeCLIClient(cfg)
	case ProviderAnthropic:
		return NewAnthropicClient(cfg)
	default:
		return NewClient()
	}
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
				return "", enhanceCredentialError(err)
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
			return "", enhanceCredentialError(err)
		}

		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	// Parse response
	var response BedrockResponse
	if err := json.Unmarshal(result.Body, &response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the content from Claude's response
	if len(response.Content) > 0 {
		fmt.Print("✔ :: Response received from Bedrock\n")
		return response.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in response")
}

// enhanceCredentialError provides helpful instructions for credential errors
func enhanceCredentialError(err error) error {
	return fmt.Errorf(`AWS credential error detected

The error suggests your AWS credentials are not properly configured or have expired.

Quick Fix - Using NEW aws login:
---------------------------------
1. Check AWS CLI version (need v2.32.0+):
   aws --version

2. Log in:
   aws login --profile my-aws

3. Add this to ~/.aws/config:

   [profile my-aws]
   login_session = arn:aws:iam::123456789012:user/yourname
   region = us-east-1

   [profile gudcommit]
   credential_process = aws configure export-credentials --profile my-aws --format process
   region = us-east-1

4. Test it:
   aws sts get-caller-identity --profile gudcommit

5. Run gudcommit (no need to set AWS_PROFILE, defaults to gudcommit):
   gudcommit

Alternative - Environment Variables:
-------------------------------------
export AWS_ACCESS_KEY_ID=your-key-id
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
gudcommit

For detailed help: golang/QUICKSTART_AWS_LOGIN.md

Original error: %v`, err)
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
