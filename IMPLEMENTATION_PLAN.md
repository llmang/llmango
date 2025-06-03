# Auto-Detection Implementation Plan: ProviderRequireParameters

## 🎯 Objective
Implement intelligent auto-detection that automatically sets `ProviderRequireParameters = true` when structured output (`ResponseFormat`) is used, ensuring only compatible providers receive structured output requests.

## 🏗️ Architecture Overview

### Core Strategy: Smart Auto-Detection
```
IF ResponseFormat is set (structured output detected)
THEN automatically set ProviderRequireParameters = true
```

### Benefits
- ✅ **Non-breaking**: Existing code continues to work unchanged
- ✅ **Automatic**: No manual configuration needed  
- ✅ **Intelligent**: Only activates when structured output is actually used
- ✅ **Clean**: Logic lives in the right place (OpenRouter package)

## 📋 Implementation Steps

### Step 1: Add Auto-Detection Function
**File**: `openrouter/openrouter.go`
**Location**: Add new method to `OpenRouterRequest`
**Function**: `autoConfigureProviderRequirements()`

```go
// autoConfigureProviderRequirements automatically sets ProviderRequireParameters
// to true when ResponseFormat is detected (structured output)
func (r *OpenRouterRequest) autoConfigureProviderRequirements() {
    if r.Parameters.ResponseFormat != nil && len(r.Parameters.ResponseFormat) > 0 {
        if r.Parameters.ProviderRequireParameters == nil {
            requireParams := true
            r.Parameters.ProviderRequireParameters = &requireParams
            log.Printf("🔧 Auto-detected structured output: setting require_parameters=true")
        }
    }
}
```

### Step 2: Integrate Auto-Detection in Request Execution
**File**: `openrouter/openrouter.go`
**Function**: `executeOpenRouterRequest` (line ~270)
**Change**: Call auto-detection before JSON marshaling

```go
func (o *OpenRouter) executeOpenRouterRequest(request *OpenRouterRequest) ([]byte, error) {
    if o.ApiKey == "" {
        return nil, errors.New("API KEY is empty in openrouter instance")
    }
    
    // Auto-configure provider requirements based on request content
    request.autoConfigureProviderRequirements()
    
    // Ensure stream is not accidentally set for this helper
    if request.Stream != nil && *request.Stream {
        return nil, errors.New("executeOpenRouterRequest is for non-streaming requests; use GenerateStreamingChatResponse for streaming")
    }

    // ... rest of existing function unchanged
}
```

### Step 3: Add Auto-Detection to Streaming Requests
**File**: `openrouter/openrouter.go`
**Function**: `GenerateStreamingChatResponse` (line ~373)
**Change**: Call auto-detection before JSON marshaling

```go
func (o *OpenRouter) GenerateStreamingChatResponse(ctx context.Context, request *OpenRouterRequest) (<-chan *StreamingChatResponse, error) {
    // Ensure stream is explicitly set to true
    if request.Stream == nil || !*request.Stream {
        stream := true
        request.Stream = &stream
    }

    // Auto-configure provider requirements based on request content
    request.autoConfigureProviderRequirements()

    // Marshal the request body to JSON
    jsonData, err := json.Marshal(request)
    // ... rest of existing function unchanged
}
```

### Step 4: Create Comprehensive Tests
**File**: `openrouter/auto_detection_test.go` (new file)
**Purpose**: Test auto-detection behavior

```go
package openrouter

import (
    "encoding/json"
    "testing"
)

func TestAutoConfigureProviderRequirements(t *testing.T) {
    tests := []struct {
        name           string
        responseFormat json.RawMessage
        existingValue  *bool
        expectedValue  *bool
        shouldLog      bool
    }{
        {
            name:           "Auto-set when ResponseFormat present and ProviderRequireParameters nil",
            responseFormat: json.RawMessage(`{"type": "json_schema"}`),
            existingValue:  nil,
            expectedValue:  &[]bool{true}[0],
            shouldLog:      true,
        },
        {
            name:           "No change when ResponseFormat present but ProviderRequireParameters already set",
            responseFormat: json.RawMessage(`{"type": "json_schema"}`),
            existingValue:  &[]bool{false}[0],
            expectedValue:  &[]bool{false}[0],
            shouldLog:      false,
        },
        {
            name:           "No change when ResponseFormat empty",
            responseFormat: nil,
            existingValue:  nil,
            expectedValue:  nil,
            shouldLog:      false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            request := &OpenRouterRequest{
                Parameters: Parameters{
                    ResponseFormat:            tt.responseFormat,
                    ProviderRequireParameters: tt.existingValue,
                },
            }

            request.autoConfigureProviderRequirements()

            if tt.expectedValue == nil {
                if request.Parameters.ProviderRequireParameters != nil {
                    t.Errorf("Expected ProviderRequireParameters to be nil, got %v", 
                        *request.Parameters.ProviderRequireParameters)
                }
            } else {
                if request.Parameters.ProviderRequireParameters == nil {
                    t.Errorf("Expected ProviderRequireParameters to be %v, got nil", *tt.expectedValue)
                } else if *request.Parameters.ProviderRequireParameters != *tt.expectedValue {
                    t.Errorf("Expected ProviderRequireParameters to be %v, got %v", 
                        *tt.expectedValue, *request.Parameters.ProviderRequireParameters)
                }
            }
        })
    }
}
```

### Step 5: Integration Tests
**File**: `openrouter/integration_test.go` (update existing)
**Purpose**: Verify end-to-end behavior

```go
func TestStructuredOutputAutoRequiresParameters(t *testing.T) {
    // Test that when we use structured output, require_parameters is automatically set
    exampleJSON := json.RawMessage(`{"name": "John", "age": 30}`)
    
    responseFormat, err := UseOpenRouterJsonFormatFromJSON(exampleJSON, "TestSchema")
    if err != nil {
        t.Fatalf("Failed to create response format: %v", err)
    }

    request := &OpenRouterRequest{
        Messages: []Message{{Role: "user", Content: "Test"}},
        Parameters: Parameters{
            ResponseFormat: responseFormat,
        },
    }

    // Simulate what happens in executeOpenRouterRequest
    request.autoConfigureProviderRequirements()

    // Verify that ProviderRequireParameters was automatically set to true
    if request.Parameters.ProviderRequireParameters == nil {
        t.Error("Expected ProviderRequireParameters to be automatically set, but it was nil")
    } else if !*request.Parameters.ProviderRequireParameters {
        t.Error("Expected ProviderRequireParameters to be true, but it was false")
    }
}
```

### Step 6: Update Documentation
**File**: `README.md` or relevant docs
**Purpose**: Document the auto-detection behavior

```markdown
## Automatic Provider Requirements

When using structured output (via `ResponseFormat`), LLMango automatically sets `require_parameters: true` to ensure only providers that support structured output receive your requests. This prevents routing to incompatible providers and ensures reliable structured responses.

This behavior is:
- **Automatic**: No configuration needed
- **Non-breaking**: Existing code continues to work
- **Smart**: Only activates when structured output is used
```

## 🧪 Testing Strategy

### Unit Tests
- Test auto-detection logic with various scenarios
- Test that existing values are preserved
- Test that nil ResponseFormat doesn't trigger auto-detection

### Integration Tests  
- Test end-to-end with actual structured output requests
- Verify that the dual-path execution system works correctly
- Test both streaming and non-streaming requests

### Manual Testing
- Test with the example app to verify real-world behavior
- Verify logging output shows auto-detection
- Test with different models and providers

## 🔄 Rollback Plan

If issues arise, the changes can be easily rolled back by:
1. Commenting out the `autoConfigureProviderRequirements()` calls
2. The auto-detection function itself is non-destructive
3. All existing functionality remains unchanged

## 📊 Success Criteria

- ✅ Auto-detection works for both streaming and non-streaming requests
- ✅ Existing code continues to work without changes
- ✅ Structured output requests only go to compatible providers
- ✅ Clear logging shows when auto-detection triggers
- ✅ All tests pass
- ✅ Example app demonstrates the functionality

## 🚀 Implementation Order

1. **Add auto-detection function** - Core logic
2. **Integrate in non-streaming requests** - Most common case
3. **Integrate in streaming requests** - Complete coverage
4. **Add unit tests** - Verify logic
5. **Add integration tests** - Verify end-to-end
6. **Test with example app** - Real-world validation
7. **Update documentation** - User awareness

This implementation ensures that the dual-path execution system automatically uses only compatible providers for structured output, making the system more reliable and user-friendly.