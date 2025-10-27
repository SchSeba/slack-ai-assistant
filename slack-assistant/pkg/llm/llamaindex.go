// Package llm provides integration with LlamaIndex RAG server for AI chat functionality.
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

// LlamaIndexClient implements the LLMClientInterface for LlamaIndex server
type LlamaIndexClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLlamaIndexClient creates a new LlamaIndex client
func NewLlamaIndexClient() Interface {
	host := os.Getenv("LLAMAINDEX_HOST")
	if host == "" {
		host = "http://localhost:5000"
	}

	return &LlamaIndexClient{
		baseURL:    host,
		httpClient: &http.Client{},
	}
}

// CreateThread generates a UUID thread slug locally (no server call needed)
func (c *LlamaIndexClient) CreateThread(project, version string) (string, error) {
	// Generate UUID locally
	threadSlug := uuid.New().String()
	fmt.Printf("Generated thread slug: %s for project=%s, version=%s\n", threadSlug, project, version)
	return threadSlug, nil
}

// SendMessageToChat sends a message to the /v1/answer endpoint
func (c *LlamaIndexClient) SendMessageToChat(project, version, threadSlug, message string) (string, error) {
	url := fmt.Sprintf("%s/v1/answer", c.baseURL)

	requestBody := map[string]interface{}{
		"project":     project,
		"version":     version,
		"thread_slug": threadSlug,
		"message":     message,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("server returned status %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		TextResponse string `json:"textResponse"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.TextResponse, nil
}

// Elaborate sends a message to the /v1/elaborate endpoint
func (c *LlamaIndexClient) Elaborate(threadSlug, message string) (string, error) {
	url := fmt.Sprintf("%s/v1/elaborate", c.baseURL)

	requestBody := map[string]interface{}{
		"thread_slug": threadSlug,
		"message":     message,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return "", fmt.Errorf("server returned status %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		return "", fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		TextResponse string `json:"textResponse"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.TextResponse, nil
}

// Inject sends content to the /v1/inject endpoint
func (c *LlamaIndexClient) Inject(project, version, message string) error {
	url := fmt.Sprintf("%s/v1/inject", c.baseURL)

	requestBody := map[string]interface{}{
		"project":     project,
		"version":     version,
		"textContent": message,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("server returned status %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
