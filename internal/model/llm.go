package model

// CommandResponse represents a structured response with command and explanation
type CommandResponse struct {
	Command     string `json:"command"`
	Details     string `json:"details"`
	ShowDetails bool   `json:"show_details"`
}

// LLMUsage tracks API usage information
type LLMUsage struct {
	Model        string
	InputTokens  int
	OutputTokens int
}
