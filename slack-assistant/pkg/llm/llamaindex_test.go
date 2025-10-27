package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLlamaIndexClient_CreateThread(t *testing.T) {
	client := &LlamaIndexClient{
		baseURL:    "http://test",
		httpClient: &http.Client{},
	}

	threadSlug, err := client.CreateThread("sriov", "4.16")
	if err != nil {
		t.Fatalf("CreateThread failed: %v", err)
	}

	if threadSlug == "" {
		t.Error("Expected non-empty thread slug")
	}
}

func TestLlamaIndexClient_SendMessageToChat(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/answer" {
			t.Errorf("Expected path /v1/answer, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request: %v", err)
		}

		if req["project"] != "sriov" || req["version"] != "4.16" {
			t.Error("Unexpected project or version in request")
		}

		response := map[string]string{
			"textResponse": "Test response",
		}
		w.Header().Set("Content-Type", "application/json")
		//nolint:errcheck // test mock
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &LlamaIndexClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	response, err := client.SendMessageToChat("sriov", "4.16", "test-thread", "test message")
	if err != nil {
		t.Fatalf("SendMessageToChat failed: %v", err)
	}

	if response != "Test response" {
		t.Errorf("Expected 'Test response', got '%s'", response)
	}
}

func TestLlamaIndexClient_SendMessageToChat_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		//nolint:errcheck // test mock
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
	}))
	defer server.Close()

	client := &LlamaIndexClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	_, err := client.SendMessageToChat("unknown", "1.0", "test-thread", "test message")
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestLlamaIndexClient_Elaborate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/elaborate" {
			t.Errorf("Expected path /v1/elaborate, got %s", r.URL.Path)
		}

		var req map[string]interface{}
		//nolint:errcheck // test mock
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req["thread_slug"] != "test-thread" {
			t.Error("Unexpected thread_slug in request")
		}

		response := map[string]string{
			"textResponse": "Elaborated response",
		}
		//nolint:errcheck // test mock
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &LlamaIndexClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	response, err := client.Elaborate("test-thread", "elaborate this")
	if err != nil {
		t.Fatalf("Elaborate failed: %v", err)
	}

	if response != "Elaborated response" {
		t.Errorf("Expected 'Elaborated response', got '%s'", response)
	}
}

func TestLlamaIndexClient_Inject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/inject" {
			t.Errorf("Expected path /v1/inject, got %s", r.URL.Path)
		}

		var req map[string]interface{}
		//nolint:errcheck // test mock
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req["project"] != "metallb" || req["version"] != "4.18" {
			t.Error("Unexpected project or version in request")
		}

		if req["textContent"] != "injected content" {
			t.Error("Unexpected textContent in request")
		}

		w.WriteHeader(http.StatusOK)
		//nolint:errcheck // test mock
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := &LlamaIndexClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	err := client.Inject("metallb", "4.18", "injected content")
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}
}

func TestLlamaIndexClient_Inject_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		//nolint:errcheck // test mock
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Bad request"})
	}))
	defer server.Close()

	client := &LlamaIndexClient{
		baseURL:    server.URL,
		httpClient: &http.Client{},
	}

	err := client.Inject("test", "1.0", "content")
	if err == nil {
		t.Error("Expected error for 400 response")
	}
}
