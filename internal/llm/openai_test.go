package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIProviderGenerate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q", got)
		}
		var req oaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if req.Model != "test-model" {
			t.Errorf("model = %q", req.Model)
		}
		if req.Stream {
			t.Error("Generate must not set stream")
		}
		_ = json.NewEncoder(w).Encode(oaResponse{
			Choices: []oaChoice{{Message: oaMessage{Role: "assistant", Content: "hello from fake"}}},
		})
	}))
	defer srv.Close()

	p := &OpenAIProvider{IDStr: "fake", BaseURL: srv.URL + "/v1", APIKey: "test-key", Model: "test-model"}
	got, err := p.Generate(context.Background(), GenerateRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got.Content != "hello from fake" {
		t.Errorf("content = %q", got.Content)
	}
	if len(got.ToolCalls) != 0 {
		t.Errorf("unexpected tool calls: %+v", got.ToolCalls)
	}
}

func TestOpenAIProviderToolCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req oaRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Tools) == 0 {
			t.Error("expected tools in request")
		}
		_ = json.NewEncoder(w).Encode(oaResponse{
			Choices: []oaChoice{{
				FinishReason: "tool_calls",
				Message: oaMessage{
					Role: "assistant",
					ToolCalls: []oaToolCall{{
						ID: "call_1", Type: "function",
						Function: oaFunction{Name: "bash", Arguments: `{"command":"echo hi"}`},
					}},
				},
			}},
		})
	}))
	defer srv.Close()

	p := &OpenAIProvider{IDStr: "fake", BaseURL: srv.URL + "/v1", APIKey: "k", Model: "m"}
	resp, err := p.Generate(context.Background(), GenerateRequest{
		Messages: []Message{{Role: "user", Content: "run it"}},
		Tools:    []ToolSchema{{Name: "bash", Description: "run", Parameters: map[string]any{"type": "object"}}},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(resp.ToolCalls) != 1 || resp.ToolCalls[0].Name != "bash" {
		t.Fatalf("tool calls = %+v", resp.ToolCalls)
	}
	if resp.ToolCalls[0].Arguments != `{"command":"echo hi"}` {
		t.Errorf("arguments = %q", resp.ToolCalls[0].Arguments)
	}
}

func TestOpenAIProviderStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		writeDelta := func(delta string) {
			b, _ := json.Marshal(oaResponse{Choices: []oaChoice{{Delta: oaMessage{Content: delta}}}})
			_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
			flusher.Flush()
		}
		writeDelta("hel")
		writeDelta("lo")
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	p := &OpenAIProvider{IDStr: "fake", BaseURL: srv.URL + "/v1", APIKey: "k", Model: "m"}
	ch, err := p.Stream(context.Background(), GenerateRequest{
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	var got string
	for ev := range ch {
		if ev.Err != nil {
			t.Fatalf("stream error: %v", ev.Err)
		}
		got += ev.Delta
		if ev.Done {
			break
		}
	}
	if got != "hello" {
		t.Errorf("streamed %q, want %q", got, "hello")
	}
}
