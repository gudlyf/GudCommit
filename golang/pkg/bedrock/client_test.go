package bedrock

import (
	"encoding/json"
	"testing"
)

func TestBedrockRequest(t *testing.T) {
	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1000,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Generate a commit message",
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled BedrockRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal request: %v", err)
	}

	if unmarshaled.AnthropicVersion != request.AnthropicVersion {
		t.Errorf("Expected AnthropicVersion %s, got %s", request.AnthropicVersion, unmarshaled.AnthropicVersion)
	}

	if unmarshaled.MaxTokens != request.MaxTokens {
		t.Errorf("Expected MaxTokens %d, got %d", request.MaxTokens, unmarshaled.MaxTokens)
	}

	if len(unmarshaled.Messages) != len(request.Messages) {
		t.Errorf("Expected %d messages, got %d", len(request.Messages), len(unmarshaled.Messages))
	}
}

func TestMessage(t *testing.T) {
	message := Message{
		Role:    "user",
		Content: "Test message content",
	}

	if message.Role != "user" {
		t.Errorf("Expected role 'user', got %s", message.Role)
	}

	if message.Content != "Test message content" {
		t.Errorf("Expected content 'Test message content', got %s", message.Content)
	}
}

func TestBedrockResponse(t *testing.T) {
	response := BedrockResponse{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "Generated commit message",
			},
		},
		Usage: Usage{
			InputTokens:  100,
			OutputTokens: 50,
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal response: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled BedrockResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(unmarshaled.Content) != len(response.Content) {
		t.Errorf("Expected %d content blocks, got %d", len(response.Content), len(unmarshaled.Content))
	}

	if unmarshaled.Usage.InputTokens != response.Usage.InputTokens {
		t.Errorf("Expected InputTokens %d, got %d", response.Usage.InputTokens, unmarshaled.Usage.InputTokens)
	}
}

func TestContentBlock(t *testing.T) {
	block := ContentBlock{
		Type: "text",
		Text: "This is a test content block",
	}

	if block.Type != "text" {
		t.Errorf("Expected type 'text', got %s", block.Type)
	}

	if block.Text != "This is a test content block" {
		t.Errorf("Expected text 'This is a test content block', got %s", block.Text)
	}
}

func TestUsage(t *testing.T) {
	usage := Usage{
		InputTokens:  150,
		OutputTokens: 75,
	}

	if usage.InputTokens != 150 {
		t.Errorf("Expected InputTokens 150, got %d", usage.InputTokens)
	}

	if usage.OutputTokens != 75 {
		t.Errorf("Expected OutputTokens 75, got %d", usage.OutputTokens)
	}
}

func TestClient(t *testing.T) {
	client := &Client{
		APIKey: "test-api-key",
		Region: "us-east-1",
	}

	if client.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey 'test-api-key', got %s", client.APIKey)
	}

	if client.Region != "us-east-1" {
		t.Errorf("Expected Region 'us-east-1', got %s", client.Region)
	}
}

func TestDefaultValues(t *testing.T) {
	if DefaultAWSRegion != "us-east-1" {
		t.Errorf("Expected DefaultAWSRegion 'us-east-1', got %s", DefaultAWSRegion)
	}
}

func TestJSONSerialization(t *testing.T) {
	// Test complete request/response cycle
	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1000,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Generate a commit message for adding user authentication",
			},
		},
	}

	// Marshal request
	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
	}

	// Unmarshal request
	var unmarshaledRequest BedrockRequest
	err = json.Unmarshal(requestJSON, &unmarshaledRequest)
	if err != nil {
		t.Errorf("Failed to unmarshal request: %v", err)
	}

	// Verify request data
	if unmarshaledRequest.AnthropicVersion != request.AnthropicVersion {
		t.Errorf("AnthropicVersion mismatch")
	}

	if unmarshaledRequest.MaxTokens != request.MaxTokens {
		t.Errorf("MaxTokens mismatch")
	}

	if len(unmarshaledRequest.Messages) != len(request.Messages) {
		t.Errorf("Messages length mismatch")
	}

	// Test response
	response := BedrockResponse{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "feat(auth): Add user authentication",
			},
		},
		Usage: Usage{
			InputTokens:  50,
			OutputTokens: 25,
		},
	}

	// Marshal response
	responseJSON, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal response: %v", err)
	}

	// Unmarshal response
	var unmarshaledResponse BedrockResponse
	err = json.Unmarshal(responseJSON, &unmarshaledResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	// Verify response data
	if len(unmarshaledResponse.Content) != len(response.Content) {
		t.Errorf("Content length mismatch")
	}

	if unmarshaledResponse.Usage.InputTokens != response.Usage.InputTokens {
		t.Errorf("InputTokens mismatch")
	}
}

// Benchmark tests
func BenchmarkBedrockRequestMarshal(b *testing.B) {
	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        1000,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Generate a commit message",
			},
		},
	}

	for i := 0; i < b.N; i++ {
		json.Marshal(request)
	}
}

func BenchmarkBedrockResponseMarshal(b *testing.B) {
	response := BedrockResponse{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: "Generated commit message",
			},
		},
		Usage: Usage{
			InputTokens:  100,
			OutputTokens: 50,
		},
	}

	for i := 0; i < b.N; i++ {
		json.Marshal(response)
	}
}
