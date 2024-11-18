package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/belingud/gptcommit/internal/config"
)

// Client represents an LLM client
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest represents a chat completion request
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

// CompletionResponse represents a chat completion response
type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// NewClient creates a new LLM client
func NewClient(cfg *config.Config) (*Client, error) {
	httpClient := &http.Client{}

	// Configure proxy if specified
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	return &Client{
		config:     cfg,
		httpClient: httpClient,
	}, nil
}

// Complete sends a completion request to the LLM API
func (c *Client) Complete(messages []Message) (string, error) {
	reqBody := CompletionRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.APIBase+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	var resp *http.Response
	var lastErr error

	// Retry logic
	for i := 0; i < c.config.Retries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil {
			break
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to send request after %d retries: %w", c.config.Retries, lastErr)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var completionResp CompletionResponse
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(completionResp.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	return completionResp.Choices[0].Message.Content, nil
}

// GenerateCommitMessage generates a commit message based on the diff
func (c *Client) GenerateCommitMessage(diff string) (string, error) {
	prompt := fmt.Sprintf(`As an AI assistant, help me generate a meaningful commit message for the following code changes:

%s

Please generate a commit message that follows these guidelines:
1. Start with a concise summary line (max 50 characters)
2. Follow with a blank line
3. Add a more detailed description if necessary
4. Use the imperative mood in the subject line
5. Do not end the subject line with a period
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how

Please respond with only the commit message, no additional explanations.`, diff)

	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that generates meaningful git commit messages.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return c.Complete(messages)
}
