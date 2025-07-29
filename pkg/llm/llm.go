package llm

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"

	anythingllm "github.com/SchSeba/anythingllm-go-sdk"
)

// LLMClient implements the LLMClientInterface
type LLMClient struct {
	apiClient *anythingllm.APIClient
}

func NewLLMClient() Interface {
	config := anythingllm.NewConfiguration()
	config.Host = os.Getenv("ANYTHINGLLM_HOST")
	config.Scheme = "http"
	config.Debug = true
	config.DefaultHeader = map[string]string{
		"Authorization": "Bearer " + os.Getenv("ANYTHINGLLM_API_KEY"),
	}
	return &LLMClient{
		apiClient: anythingllm.NewAPIClient(config),
	}
}

func (c *LLMClient) CreateThread(project, version string) (string, error) {
	slug := project
	if version != "" {
		version = strings.ReplaceAll(version, ".", "-dot-")
		slug = fmt.Sprintf("%s-%s", project, version)
	}

	// Check if the slug exist
	workspaceInfoRequest := c.apiClient.WorkspacesAPI.V1WorkspaceSlugGet(context.Background(), slug)
	workspaceInfo, response, err := workspaceInfoRequest.Execute()
	if err != nil {
		fmt.Printf("❌ Failed to get workspace info: %v\n", err)
		return "", err
	}
	fmt.Printf("Workspace info: %+v\n", workspaceInfo)

	request := c.apiClient.WorkspaceThreadsAPI.V1WorkspaceSlugThreadNewPost(context.Background(), slug)
	slugThreadInfo, response, err := request.Execute()
	fmt.Printf("HTTP Response Status: %s\n", response.Status)
	if err != nil {
		return "", err
	}

	threadResponse, err := ConvertMapToWorkspaceThread(slugThreadInfo["thread"])
	if err != nil {
		return "", fmt.Errorf("failed to convert response to struct: %w", err)
	}
	fmt.Printf("Thread created: ID=%d, Slug=%s, Name=%s\n",
		threadResponse.ID, threadResponse.Slug, threadResponse.Name)

	return threadResponse.Slug, nil
}

func (c *LLMClient) SendMessageToChat(project, version, threadSlug, message string) (string, error) {
	slug := project
	if version != "" {
		version = strings.ReplaceAll(version, ".", "-dot-")
		slug = fmt.Sprintf("%s-%s", project, version)
	}

	return c.sendMessageToChatWithMode(slug, threadSlug, message, "query")
}

func (c *LLMClient) Elaborate(threadSlug, message string) (string, error) {
	return c.sendMessageToChatWithMode("elaborate", threadSlug, message, "chat")
}

func (c *LLMClient) Inject(project, version, message string) error {
	version = strings.ReplaceAll(version, ".", "-dot-")
	wokerspace := fmt.Sprintf("%s-%s", project, version)
	request := c.apiClient.DocumentsAPI.V1DocumentRawTextPost(context.Background()).Body(map[string]interface{}{
		"textContent":     message,
		"addToWorkspaces": wokerspace,
		"metadata": map[string]interface{}{
			"title": fmt.Sprintf("Document-%d", rand.Intn(1000000)),
		},
	})
	documentInjectInfo, response, err := request.Execute()
	if err != nil {
		return fmt.Errorf("failed to inject messages: %w", err)
	}
	fmt.Printf("HTTP Response Status: %s\n", response.Status)
	fmt.Printf("Document inject info: %+v\n", documentInjectInfo)
	return nil
}

func (c *LLMClient) sendMessageToChatWithMode(slug, threadSlug, message, mode string) (string, error) {
	request := c.apiClient.WorkspaceThreadsAPI.V1WorkspaceSlugThreadThreadSlugChatPost(
		context.Background(),
		slug,
		threadSlug,
	).V1WorkspaceSlugThreadThreadSlugChatPostRequest(anythingllm.V1WorkspaceSlugThreadThreadSlugChatPostRequest{
		Message: message,
		Mode:    &mode,
		UserId:  *anythingllm.NewNullableInt32(anythingllm.PtrInt32(2)),
	})
	chatInfo, response, err := request.Execute()
	fmt.Printf("HTTP Response Status: %s\n", response.Status)
	if err != nil {
		return "", err
	}
	fmt.Printf("Chat response: %+v\n", chatInfo)
	chatResponse, err := ConvertMapToChatResponse(chatInfo)
	if err != nil {
		return "", fmt.Errorf("failed to convert response to struct: %w", err)
	}
	fmt.Printf("Chat response: %+v\n", chatResponse)
	return chatResponse.TextResponse, nil
}
