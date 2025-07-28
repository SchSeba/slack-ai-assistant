package llm

import (
	"encoding/json"
	"fmt"
)

// Interface defines the interface for LLM client operations
type Interface interface {
	CreateThread(project, version string) (string, error)
	SendMessageToChat(project, version, threadSlug, message string) (string, error)
	Elaborate(threadSlug, message string) (string, error)
	Inject(project, version, message string) error
}

// WorkspaceThreadResponse represents the response from creating a new thread
type WorkspaceThreadResponse struct {
	ID          int64  `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	WorkspaceID string `json:"workspaceId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// ChatMessage represents a chat message in a thread
type ChatBody struct {
	Message string `json:"message"`
	UserId  int64  `json:"userId"`
	Mode    string `json:"mode"`
}

type ChatResponse struct {
	ID           string `json:"id"`
	TextResponse string `json:"textResponse"`
	// Sources      []string `json:"sources"`
}

// ConvertMapToWorkspaceThread converts map[string]interface{} to WorkspaceThreadResponse
func ConvertMapToWorkspaceThread(data interface{}) (*WorkspaceThreadResponse, error) {
	// You can use a JSON marshaling approach for type safety
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	var thread WorkspaceThreadResponse
	if err := json.Unmarshal(jsonData, &thread); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to struct: %w", err)
	}

	return &thread, nil
}

func ConvertMapToChatResponse(data interface{}) (*ChatResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	var chat ChatResponse
	if err := json.Unmarshal(jsonData, &chat); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to struct: %w", err)
	}

	return &chat, nil
}
