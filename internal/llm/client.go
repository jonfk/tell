package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jonfk/tell/internal/config"
)

// Client represents an LLM API client
type Client struct {
	config *config.Config
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicRequest represents a request to the Anthropic API
type AnthropicRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

// AnthropicResponse represents a response from the Anthropic API
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// CommandResponse represents a structured response with command and explanation
type CommandResponse struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
}

// NewClient creates a new LLM client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

// GenerateCommand generates a shell command from a natural language prompt
func (c *Client) GenerateCommand(prompt string, includeContext bool) (*CommandResponse, error) {
	// TODO: Implement context gathering if includeContext is true
	
	// Build system prompt
	systemPrompt := buildSystemPrompt(c.config, includeContext)

	// Create messages
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	slog.Debug("Sending request to LLM", 
		"model", c.config.LLMModel,
		"includeContext", includeContext,
		"promptLength", len(prompt))

	// Call Anthropic API
	apiResp, err := c.callAnthropicAPI(messages)
	if err != nil {
		slog.Error("LLM API request failed", "error", err)
		return nil, err
	}

	slog.Debug("Received response from LLM", 
		"inputTokens", apiResp.Usage.InputTokens,
		"outputTokens", apiResp.Usage.OutputTokens)

	// Parse the response
	if len(apiResp.Content) == 0 {
		slog.Error("Empty response from API")
		return nil, fmt.Errorf("empty response from API")
	}

	// For now, just return the raw text as both command and explanation
	// TODO: Implement proper structured output parsing
	text := apiResp.Content[0].Text

	// Simple parsing - this should be improved with proper structured output
	parts := strings.SplitN(text, "\n\n", 2)

	resp := &CommandResponse{}
	if len(parts) > 1 {
		resp.Command = strings.TrimSpace(parts[0])
		resp.Explanation = strings.TrimSpace(parts[1])
	} else {
		resp.Command = strings.TrimSpace(text)
		resp.Explanation = "No explanation provided"
	}

	slog.Debug("Parsed command response", 
		"commandLength", len(resp.Command),
		"explanationLength", len(resp.Explanation))

	return resp, nil
}

// callAnthropicAPI calls the Anthropic API with the given messages
func (c *Client) callAnthropicAPI(messages []Message) (*AnthropicResponse, error) {
	// Create the request body
	reqBody := AnthropicRequest{
		Model:     c.config.LLMModel,
		Messages:  messages,
		MaxTokens: 1000,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp, nil
}
