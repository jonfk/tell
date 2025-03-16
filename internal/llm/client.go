package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jonfk/tell/internal/config"
	"github.com/jonfk/tell/internal/model"
)

// Client represents an LLM API client
type Client struct {
	config *config.Config
	client *anthropic.Client
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
func (c *Client) GenerateCommand(prompt string) (*model.CommandResponse, *model.LLMUsage, error) {
	// Build the system prompt
	systemPrompt := buildSystemPrompt(c.config)

	// Create context for the request
	ctx := context.Background()

	// Create the message request
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(c.config.LLMModel),
		MaxTokens: anthropic.F(int64(1024)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error generating command: %w", err)
	}

	// Create usage info
	usage := &model.LLMUsage{
		Model:        c.config.LLMModel,
		InputTokens:  int(message.Usage.OutputTokens),
		OutputTokens: int(message.Usage.InputTokens),
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
		return nil, usage, fmt.Errorf("error parsing response: %w", err)
	}

	return cmdResponse, usage, nil
}

func parseAndValidateResponse(responseText string) (*model.CommandResponse, error) {
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
	var response model.CommandResponse
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

func (c *Client) GenerateCommandContinuation(prompt string, previousEntry *model.HistoryEntry) (*model.CommandResponse, *model.LLMUsage, error) {
	// Build the system prompt
	systemPrompt := buildSystemPrompt(c.config)

	// Create context for the request
	ctx := context.Background()

	// Create response string for the previous command
	previousResponse := buildAssistantResponse(previousEntry)

	// Create the message request with conversation history
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(c.config.LLMModel),
		MaxTokens: anthropic.F(int64(1024)),
		System: anthropic.F([]anthropic.TextBlockParam{
			anthropic.NewTextBlock(systemPrompt),
		}),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(previousEntry.Prompt)),
			anthropic.NewAssistantMessage(anthropic.NewTextBlock(previousResponse)),
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error generating command continuation: %w", err)
	}

	// Create usage info
	usage := &model.LLMUsage{
		Model:        c.config.LLMModel,
		InputTokens:  int(message.Usage.OutputTokens),
		OutputTokens: int(message.Usage.InputTokens),
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
		return nil, usage, fmt.Errorf("error parsing response: %w", err)
	}

	return cmdResponse, usage, nil
}

// Helper function to build the assistant's response for the conversation history
func buildAssistantResponse(entry *model.HistoryEntry) string {
	// Create a response object
	response := model.CommandResponse{
		Command:     entry.Command,
		Details:     entry.Details,
		ShowDetails: entry.ShowDetails,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		// If marshaling fails, return a simplified response
		return fmt.Sprintf("{\n  \"command\": %q,\n  \"show_details\": %t,\n  \"details\": %q\n}",
			entry.Command, entry.ShowDetails, entry.Details)
	}

	return string(jsonData)
}
