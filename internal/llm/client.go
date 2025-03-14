package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/anthropic-ai/anthropic-sdk-go"
	"github.com/jonfk/tell/internal/config"
)

// Client represents an LLM API client
type Client struct {
	config *config.Config
	client *anthropic.Client
}

// CommandResponse represents a structured response with command and explanation
type CommandResponse struct {
	Command   string `json:"command"`
	ShortDesc string `json:"short_description"`
	LongDesc  string `json:"long_description"`
}

// parseXMLOutput parses the XML output from the LLM response
func parseXMLOutput(text string) (*CommandResponse, error) {
	// Check if the response contains the expected XML tags
	if !strings.Contains(text, "<output>") || 
	   !strings.Contains(text, "</output>") {
		return nil, fmt.Errorf("response does not contain <output> tags")
	}
	
	// Extract the content between <output> tags
	outputStart := strings.Index(text, "<output>") + len("<output>")
	outputEnd := strings.Index(text, "</output>")
	if outputStart == -1 || outputEnd == -1 || outputStart >= outputEnd {
		return nil, fmt.Errorf("invalid <output> tags in response")
	}
	
	output := strings.TrimSpace(text[outputStart:outputEnd])
	
	// Initialize response
	response := &CommandResponse{}
	
	// Extract command (required)
	commandStart := strings.Index(output, "<command>") + len("<command>")
	commandEnd := strings.Index(output, "</command>")
	if commandStart == -1 || commandEnd == -1 || commandStart >= commandEnd {
		return nil, fmt.Errorf("missing or invalid <command> tags in response")
	}
	response.Command = strings.TrimSpace(output[commandStart:commandEnd])
	
	// Extract short description
	shortStart := strings.Index(output, "<short>") + len("<short>")
	shortEnd := strings.Index(output, "</short>")
	if shortStart != -1 && shortEnd != -1 && shortStart < shortEnd {
		response.ShortDesc = strings.TrimSpace(output[shortStart:shortEnd])
	} else {
		return nil, fmt.Errorf("missing or invalid <short> tags in response")
	}
	
	// Extract long explanation
	longStart := strings.Index(output, "<long>") + len("<long>")
	longEnd := strings.Index(output, "</long>")
	if longStart != -1 && longEnd != -1 && longStart < longEnd {
		response.LongDesc = strings.TrimSpace(output[longStart:longEnd])
	} else {
		return nil, fmt.Errorf("missing or invalid <long> tags in response")
	}
	
	return response, nil
}

// NewClient creates a new LLM client
func NewClient(cfg *config.Config) *Client {
	client := anthropic.NewClient(anthropic.WithAPIKey(cfg.AnthropicAPIKey))
	
	return &Client{
		config: cfg,
		client: client,
	}
}

// GenerateCommand generates a shell command from a natural language prompt
func (c *Client) GenerateCommand(prompt string, includeContext bool) (*CommandResponse, error) {
	// Build system prompt
	systemPrompt := buildSystemPrompt(c.config, includeContext)

	slog.Debug("Sending request to LLM", 
		"model", c.config.LLMModel,
		"includeContext", includeContext,
		"promptLength", len(prompt))

	// Create message request using the SDK
	req := &anthropic.MessageRequest{
		Model:     c.config.LLMModel,
		System:    systemPrompt,
		MaxTokens: 1000,
		Messages: []anthropic.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	// Call Anthropic API using the SDK
	resp, err := c.client.Messages(context.Background(), req)
	if err != nil {
		slog.Error("LLM API request failed", "error", err)
		return nil, err
	}

	slog.Debug("Received response from LLM", 
		"inputTokens", resp.Usage.InputTokens,
		"outputTokens", resp.Usage.OutputTokens)

	// Parse the response
	if len(resp.Content) == 0 {
		slog.Error("Empty response from API")
		return nil, fmt.Errorf("empty response from API")
	}

	// Get the raw text from the response
	var text string
	for _, content := range resp.Content {
		if content.Type == "text" {
			text = content.Text
			break
		}
	}
	
	if text == "" {
		return nil, fmt.Errorf("no text content in response")
	}
	
	// Parse the XML output - no fallback mechanism
	cmdResp, err := parseXMLOutput(text)
	if err != nil {
		slog.Error("Failed to parse XML output from LLM response", "error", err, "response", text)
		return nil, fmt.Errorf("invalid response format: %w", err)
	}

	slog.Debug("Parsed command response", 
		"commandLength", len(cmdResp.Command),
		"shortDescLength", len(cmdResp.ShortDesc),
		"longDescLength", len(cmdResp.LongDesc))

	return cmdResp, nil
}

