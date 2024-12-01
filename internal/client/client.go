package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"gptcomet/pkg/types"
)

// Client represents an LLM client
type Client struct {
	config     *types.ClientConfig
	httpClient *http.Client
}

// New creates a new LLM client
func New(config *types.ClientConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// Chat sends a chat completion request
func (c *Client) Chat(messages []types.Message) (*types.CompletionResponse, error) {
	req := &types.CompletionRequest{
		Model:    c.config.Model,
		Messages: messages,
	}

	var resp *types.CompletionResponse
	var err error

	for i := 0; i <= c.config.Retries; i++ {
		resp, err = c.sendRequest(req)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed after %d retries: %w", c.config.Retries, err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from model")
	}

	return resp, nil
}

func (c *Client) sendRequest(req *types.CompletionRequest) (*types.CompletionResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if c.config.Debug {
		fmt.Printf("Debug: Sending request to provider `%s` \n", c.config.Provider)
	}

	// Trim trailing slash from APIBase
	apiBase := strings.TrimRight(c.config.APIBase, "/")
	apiURL := fmt.Sprintf("%s/chat/completions", apiBase)

	// Create HTTP client with proxy if configured
	client := c.httpClient
	if c.config.Proxy != "" {
		proxyURL, err := neturl.Parse(c.config.Proxy)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
			Timeout: c.httpClient.Timeout,
		}
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	for k, v := range c.config.ExtraHeaders {
		httpReq.Header.Set(k, v)
	}

	if c.config.Debug {
		fmt.Printf("Debug: Request headers:\n")
		for k, v := range httpReq.Header {
			if k != "Authorization" {
				fmt.Printf("%s: %v\n", k, v)
			}
		}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		if c.config.Debug {
			fmt.Printf("Debug: Request failed: %v\n", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.config.Debug {
		fmt.Printf("Debug: Response status: %d\n", resp.StatusCode)
		fmt.Printf("Debug: Response body:\n%s\n", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, body)
	}

	var completionResp types.CompletionResponse
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &completionResp, nil
}

// GenerateCommitMessage generates a commit message for the given diff
func (c *Client) GenerateCommitMessage(diff string, prompt string) (string, error) {
	if c.config.Debug {
		fmt.Printf("Debug: Generating commit message for staged diff\n")
	}

	messages := []types.Message{
		{
			Role:    "system",
			Content: prompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("{{ placeholder }}\n\n%s", diff),
		},
	}

	resp, err := c.Chat(messages)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
