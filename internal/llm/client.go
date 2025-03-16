package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jonfk/tell/internal/config"
)

// Client represents an LLM API client
type Client struct {
	config *config.Config
	client *anthropic.Client
}

// CommandResponse represents a structured response with command and explanation
type CommandResponse struct {
	Command     string `json:"command"`
	Details     string `json:"details"`
	ShowDetails bool   `json:"show_details"`
}

// NewClient creates a new LLM client
func NewClient(cfg *config.Config) *Client {
	// Create new client using the current SDK pattern
	client := anthropic.NewClient(
		option.WithAPIKey(cfg.AnthropicAPIKey),
	)

	return &Client{
		config: cfg,
		client: client,
	}
}

// GenerateCommand generates a shell command from a natural language prompt
func (c *Client) GenerateCommand(prompt string) (*CommandResponse, error) {
	// Build the system prompt
	systemPrompt := buildSystemPrompt(c.config)

	// Create context for the request
	ctx := context.Background()

	// Create the message request
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.ModelClaude3_7SonnetLatest),
		MaxTokens: anthropic.F(int64(1024)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("error generating command: %w", err)
	}

	// Extract the text content from the assistant's response
	var responseText string
	for _, content := range message.Content {
		if content.Type == anthropic.ContentBlockTypeText {
			responseText += content.Text
		}
	}

	// Parse the JSON output
	cmdResponse, err := parseAndValidateResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return cmdResponse, nil
}

func parseAndValidateResponse(responseText string) (*CommandResponse, error) {
	// Try to find JSON content in the response
	// Look for the first '{' and the last '}'
	startIdx := strings.Index(responseText, "{")
	endIdx := strings.LastIndex(responseText, "}")

	if startIdx == -1 || endIdx == -1 || endIdx <= startIdx {
		return nil, fmt.Errorf("could not find valid JSON in response: %s", responseText)
	}

	// Extract the JSON part of the response
	jsonStr := responseText[startIdx : endIdx+1]

	// Parse the JSON
	var response CommandResponse
	err := json.Unmarshal([]byte(jsonStr), &response)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w, response: %s", err, jsonStr)
	}

	// Validate the parsed response
	if response.Command == "" {
		return nil, fmt.Errorf("command is empty in response: %s", jsonStr)
	}

	return &response, nil
}
