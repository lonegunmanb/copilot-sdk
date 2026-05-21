package copilot

import (
	"encoding/json"
	"testing"
	"time"
)

func TestPermissionRequestResultKind_Constants(t *testing.T) {
	tests := []struct {
		name     string
		kind     PermissionRequestResultKind
		expected string
	}{
		{"Approved", PermissionRequestResultKindApproved, "approve-once"},
		{"Rejected", PermissionRequestResultKindRejected, "reject"},
		{"UserNotAvailable", PermissionRequestResultKindUserNotAvailable, "user-not-available"},
		{"NoResult", PermissionRequestResultKindNoResult, "no-result"},
		// Deprecated aliases
		{"DeprecatedDeniedInteractivelyByUser", PermissionRequestResultKindDeniedInteractivelyByUser, "reject"},
		{"DeprecatedDeniedCouldNotRequestFromUser", PermissionRequestResultKindDeniedCouldNotRequestFromUser, "user-not-available"},
		{"DeprecatedDeniedByRules", PermissionRequestResultKindDeniedByRules, "user-not-available"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.kind) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.kind))
			}
		})
	}
}

func TestPermissionRequestResultKind_CustomValue(t *testing.T) {
	custom := PermissionRequestResultKind("custom-kind")
	if string(custom) != "custom-kind" {
		t.Errorf("expected %q, got %q", "custom-kind", string(custom))
	}
}

func TestPermissionRequestResult_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		kind PermissionRequestResultKind
	}{
		{"Approved", PermissionRequestResultKindApproved},
		{"DeniedByRules", PermissionRequestResultKindDeniedByRules},
		{"DeniedCouldNotRequestFromUser", PermissionRequestResultKindDeniedCouldNotRequestFromUser},
		{"DeniedInteractivelyByUser", PermissionRequestResultKindDeniedInteractivelyByUser},
		{"NoResult", PermissionRequestResultKind("no-result")},
		{"Custom", PermissionRequestResultKind("custom")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := PermissionRequestResult{Kind: tt.kind}
			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var decoded PermissionRequestResult
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if decoded.Kind != tt.kind {
				t.Errorf("expected kind %q, got %q", tt.kind, decoded.Kind)
			}
		})
	}
}

func TestPermissionRequestResult_JSONDeserialize(t *testing.T) {
	jsonStr := `{"kind":"reject"}`
	var result PermissionRequestResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result.Kind != PermissionRequestResultKindRejected {
		t.Errorf("expected %q, got %q", PermissionRequestResultKindRejected, result.Kind)
	}
}

func TestPermissionRequestResult_JSONSerialize(t *testing.T) {
	result := PermissionRequestResult{Kind: PermissionRequestResultKindApproved}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"kind":"approve-once"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestPingResponse_UnmarshalJSON_NumericTimestamp(t *testing.T) {
	const raw = `{"message":"pong","timestamp":1779352370134,"protocolVersion":3}`
	var resp PingResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := resp.Timestamp.UnixMilli(), int64(1779352370134); got != want {
		t.Fatalf("Timestamp got %d, want %d", got, want)
	}
	if resp.Message != "pong" {
		t.Fatalf("Message got %q, want pong", resp.Message)
	}
	if resp.ProtocolVersion == nil || *resp.ProtocolVersion != 3 {
		t.Fatalf("ProtocolVersion got %v, want 3", resp.ProtocolVersion)
	}
}

func TestPingResponse_UnmarshalJSON_ISO8601Timestamp(t *testing.T) {
	const raw = `{"message":"pong","timestamp":"2026-05-21T08:29:54.042Z","protocolVersion":3}`
	var resp PingResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := time.Date(2026, 5, 21, 8, 29, 54, 42_000_000, time.UTC)
	if !resp.Timestamp.Equal(want) {
		t.Fatalf("Timestamp got %s, want %s", resp.Timestamp, want)
	}
}

func TestPingResponse_UnmarshalJSON_StringifiedEpoch(t *testing.T) {
	const raw = `{"message":"pong","timestamp":"1779352370134","protocolVersion":3}`
	var resp PingResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := resp.Timestamp.UnixMilli(), int64(1779352370134); got != want {
		t.Fatalf("Timestamp got %d, want %d", got, want)
	}
}

func TestPingResponse_UnmarshalJSON_RejectsGarbage(t *testing.T) {
	const raw = `{"message":"pong","timestamp":"not-a-date","protocolVersion":3}`
	var resp PingResponse
	if err := json.Unmarshal([]byte(raw), &resp); err == nil {
		t.Fatalf("expected error for garbage timestamp, got %+v", resp)
	}
}

func TestProviderConfig_JSONIncludesHeaders(t *testing.T) {
	config := ProviderConfig{
		BaseURL: "https://example.com/provider",
		Headers: map[string]string{"Authorization": "Bearer provider-token"},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal provider config: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal provider config: %v", err)
	}

	if decoded["baseUrl"] != "https://example.com/provider" {
		t.Fatalf("expected baseUrl to round-trip, got %v", decoded["baseUrl"])
	}
	headers, ok := decoded["headers"].(map[string]any)
	if !ok {
		t.Fatalf("expected headers object, got %T", decoded["headers"])
	}
	if headers["Authorization"] != "Bearer provider-token" {
		t.Fatalf("expected Authorization header, got %v", headers["Authorization"])
	}
}

func TestSessionSendRequest_JSONIncludesRequestHeaders(t *testing.T) {
	req := sessionSendRequest{
		SessionID:      "session-1",
		Prompt:         "hello",
		RequestHeaders: map[string]string{"Authorization": "Bearer turn-token"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal session send request: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal session send request: %v", err)
	}

	if decoded["prompt"] != "hello" {
		t.Fatalf("expected prompt to round-trip, got %v", decoded["prompt"])
	}
	headers, ok := decoded["requestHeaders"].(map[string]any)
	if !ok {
		t.Fatalf("expected requestHeaders object, got %T", decoded["requestHeaders"])
	}
	if headers["Authorization"] != "Bearer turn-token" {
		t.Fatalf("expected Authorization header, got %v", headers["Authorization"])
	}
}

func TestProviderConfig_JSONIncludesAllFields(t *testing.T) {
	cfg := ProviderConfig{
		BaseURL:         "https://example.com/provider",
		APIKey:          "test-key",
		Headers:         map[string]string{"Authorization": "Bearer provider-token"},
		ModelID:         "gpt-4o",
		WireModel:       "my-finetune-v3",
		MaxInputTokens:  100000,
		MaxOutputTokens: 4096,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal ProviderConfig: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ProviderConfig: %v", err)
	}

	if decoded["baseUrl"] != "https://example.com/provider" {
		t.Errorf("expected baseUrl to round-trip, got %v", decoded["baseUrl"])
	}
	if decoded["modelId"] != "gpt-4o" {
		t.Errorf("expected modelId 'gpt-4o', got %v", decoded["modelId"])
	}
	if decoded["wireModel"] != "my-finetune-v3" {
		t.Errorf("expected wireModel 'my-finetune-v3', got %v", decoded["wireModel"])
	}
	if decoded["maxPromptTokens"] != float64(100000) {
		t.Errorf("expected maxPromptTokens 100000, got %v", decoded["maxPromptTokens"])
	}
	if decoded["maxOutputTokens"] != float64(4096) {
		t.Errorf("expected maxOutputTokens 4096, got %v", decoded["maxOutputTokens"])
	}
	headers, ok := decoded["headers"].(map[string]any)
	if !ok {
		t.Fatalf("expected headers object, got %T", decoded["headers"])
	}
	if headers["Authorization"] != "Bearer provider-token" {
		t.Errorf("expected Authorization header, got %v", headers["Authorization"])
	}
}

func TestProviderConfig_JSONOmitsUnsetTokenFields(t *testing.T) {
	cfg := ProviderConfig{BaseURL: "https://example.com/provider"}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal ProviderConfig: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ProviderConfig: %v", err)
	}

	for _, field := range []string{"modelId", "wireModel", "maxPromptTokens", "maxOutputTokens", "headers"} {
		if _, present := decoded[field]; present {
			t.Errorf("expected %q to be omitted when unset, got %v", field, decoded[field])
		}
	}
}

func TestCustomAgentConfig_JSONIncludesModel(t *testing.T) {
	cfg := CustomAgentConfig{
		Name:   "model-agent",
		Prompt: "You are a model agent.",
		Model:  "claude-haiku-4.5",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal CustomAgentConfig: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CustomAgentConfig: %v", err)
	}

	if decoded["model"] != "claude-haiku-4.5" {
		t.Errorf("expected model 'claude-haiku-4.5', got %v", decoded["model"])
	}
	if decoded["name"] != "model-agent" {
		t.Errorf("expected name 'model-agent', got %v", decoded["name"])
	}
}

func TestCustomAgentConfig_JSONOmitsModelWhenEmpty(t *testing.T) {
	cfg := CustomAgentConfig{
		Name:   "no-model-agent",
		Prompt: "You are an agent without a model.",
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal CustomAgentConfig: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CustomAgentConfig: %v", err)
	}

	if _, present := decoded["model"]; present {
		t.Errorf("expected model to be omitted when empty, got %v", decoded["model"])
	}
}
