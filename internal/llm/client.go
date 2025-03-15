package llm

import (
	"context"
	"encoding/xml"
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
	Command   string `xml:"command" json:"command"`
	ShortDesc string `xml:"short" json:"short_description"`
	LongDesc  string `xml:"long" json:"long_description"`
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
func (c *Client) GenerateCommand(prompt string, includeContext bool) (*CommandResponse, error) {
	// Build the system prompt
	systemPrompt := buildSystemPrompt(c.config, includeContext)

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

	// Parse the XML output
	cmdResponse, err := parseXMLOutput(responseText)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return cmdResponse, nil
}

// parseXMLOutput parses the XML output from the LLM response
func parseXMLOutput(text string) (*CommandResponse, error) {
	// Look for <output> tags in the text
	startTag := "<output>"
	endTag := "</output>"

	startIndex := strings.Index(text, startTag)
	endIndex := strings.Index(text, endTag)

	if startIndex == -1 || endIndex == -1 || endIndex <= startIndex {
		return nil, fmt.Errorf("output XML tags not found or malformed in response")
	}

	// Extract the XML content including the tags
	xmlContent := text[startIndex : endIndex+len(endTag)]

	// Parse the XML
	var response CommandResponse
	err := xml.Unmarshal([]byte(xmlContent), &response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	// Validate that we got the required fields
	if response.Command == "" {
		return nil, fmt.Errorf("command field is empty in the response")
	}

	return &response, nil
}
