package llmangoagents

import (
	"fmt"
	"time"

	"github.com/llmang/llmango/openrouter"
)

// MockOpenRouter is a mock implementation for testing
type MockOpenRouter struct {
	// Configurable responses for testing
	DefaultResponse string
	ResponseMap     map[string]string // Map model names to specific responses
	CallCount       int
	LastRequest     *openrouter.OpenRouterRequest
	ShouldError     bool
	ErrorMessage    string
}

// NewMockOpenRouter creates a new mock OpenRouter with default settings
func NewMockOpenRouter() *MockOpenRouter {
	return &MockOpenRouter{
		DefaultResponse: "This is a mock response from the test LLM.",
		ResponseMap:     make(map[string]string),
		CallCount:       0,
		ShouldError:     false,
	}
}

// SetResponse sets a specific response for a given model
func (m *MockOpenRouter) SetResponse(model, response string) {
	m.ResponseMap[model] = response
}

// SetDefaultResponse sets the default response for all models
func (m *MockOpenRouter) SetDefaultResponse(response string) {
	m.DefaultResponse = response
}

// SetError configures the mock to return an error
func (m *MockOpenRouter) SetError(shouldError bool, errorMessage string) {
	m.ShouldError = shouldError
	m.ErrorMessage = errorMessage
}

// GenerateNonStreamingChatResponse implements the method used by the agent system
func (m *MockOpenRouter) GenerateNonStreamingChatResponse(req *openrouter.OpenRouterRequest) (*openrouter.NonStreamingChatResponse, error) {
	m.CallCount++
	m.LastRequest = req

	// Return error if configured to do so
	if m.ShouldError {
		return nil, fmt.Errorf("%s", m.ErrorMessage)
	}

	// Get response based on model
	response := m.DefaultResponse
	if req.Model != nil {
		if modelResponse, exists := m.ResponseMap[*req.Model]; exists {
			response = modelResponse
		}
	}

	// Create mock response
	mockResponse := &openrouter.NonStreamingChatResponse{
		OpenRouterBaseResponse: openrouter.OpenRouterBaseResponse{
			ID:      fmt.Sprintf("mock-completion-%d", m.CallCount),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   *req.Model,
			Usage: &openrouter.ResponseUsage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		},
		Choices: []openrouter.NonStreamingChatChoice{
			{
				Message: openrouter.ResponseMessage{
					Role:    "assistant",
					Content: &response,
				},
				BaseChoice: openrouter.BaseChoice{
					FinishReason: func() *string { s := "stop"; return &s }(),
				},
			},
		},
	}

	return mockResponse, nil
}

// CreateTestSystemManager creates a test system manager without OpenRouter for unit testing
func CreateTestSystemManager() (*AgentSystemManager, error) {
	// Get test configuration
	testConfig := GetTestConfig()

	// Create the agent system manager
	asm, err := CreateAgentSystemManager(testConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent system manager: %w", err)
	}

	// Initialize global key bank if needed
	if asm.GlobalKeyBank == nil {
		asm.GlobalKeyBank = make(map[string]string)
	}

	return asm, nil
}

// CreateTestSystemWithMockOpenRouter creates a test system with a real OpenRouter that can be mocked
// This is a helper for integration testing where you want to test the full system
func CreateTestSystemWithMockOpenRouter(apiKey string) (*AgentSystemManager, error) {
	// Create real OpenRouter
	router, err := openrouter.CreateOpenRouter(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenRouter: %w", err)
	}

	// Get test configuration
	testConfig := GetTestConfig()

	// Create the agent system manager
	asm, err := CreateAgentSystemManager(testConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent system manager: %w", err)
	}

	// Set the OpenRouter
	asm.Openrouter = router

	// Initialize global key bank if needed
	if asm.GlobalKeyBank == nil {
		asm.GlobalKeyBank = make(map[string]string)
	}

	return asm, nil
}
