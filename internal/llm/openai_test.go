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

// TestOpenAIProviderStreamToolCalls proves Stream accumulates tool calls that arrive
// fragmented across many deltas (id + name once, arguments in pieces) and surfaces
// them on the final Done event.
func TestOpenAIProviderStreamToolCalls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		write := func(v any) {
			b, _ := json.Marshal(v)
			_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
			flusher.Flush()
		}
		tc := func(id, name, args string) oaResponse {
			return oaResponse{Choices: []oaChoice{{Delta: oaMessage{ToolCalls: []oaToolCall{
				{Type: "function", ID: id, Function: oaFunction{Name: name, Arguments: args}},
			}}}}}
		}
		write(tc("call_1", "bash", ""))
		write(tc("", "", `{"command":`))
		write(tc("", "", `"echo hi"}`))
		write(oaResponse{Choices: []oaChoice{{FinishReason: "tool_calls"}}})
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	p := &OpenAIProvider{IDStr: "fake", BaseURL: srv.URL + "/v1", APIKey: "k", Model: "m"}
	ch, err := p.Stream(context.Background(), GenerateRequest{Messages: []Message{{Role: "user", Content: "run echo"}}})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	var got []ToolCall
	var done bool
	for ev := range ch {
		if ev.Err != nil {
			t.Fatalf("stream error: %v", ev.Err)
		}
		if ev.Done {
			done = true
			got = ev.ToolCalls
			break
		}
	}
	if !done {
		t.Fatal("stream did not signal Done")
	}
	if len(got) != 1 {
		t.Fatalf("tool calls = %d, want 1 (%+v)", len(got), got)
	}
	if got[0].ID != "call_1" || got[0].Name != "bash" || got[0].Arguments != `{"command":"echo hi"}` {
		t.Errorf("accumulated tool call = %+v", got[0])
	}
}

// TestOpenAIProviderStreamStripsThink proves Stream removes a reasoning think block
// from the visible deltas, even when the delimiters are split across chunk
// boundaries (content is emitted in 3-byte slices).
func TestOpenAIProviderStreamStripsThink(t *testing.T) {
	full := thinkOpen + "reasoning that should be hidden" + thinkClose + "The answer is 42."
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		for i := 0; i < len(full); i += 3 {
			end := i + 3
			if end > len(full) {
				end = len(full)
			}
			b, _ := json.Marshal(oaResponse{Choices: []oaChoice{{Delta: oaMessage{Content: full[i:end]}}}})
			_, _ = w.Write([]byte("data: " + string(b) + "\n\n"))
			flusher.Flush()
		}
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
		flusher.Flush()
	}))
	defer srv.Close()

	p := &OpenAIProvider{IDStr: "fake", BaseURL: srv.URL + "/v1", APIKey: "k", Model: "m"}
	ch, err := p.Stream(context.Background(), GenerateRequest{Messages: []Message{{Role: "user", Content: "q"}}})
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
	if got != "The answer is 42." {
		t.Errorf("streamed %q, want %q (think block must be stripped)", got, "The answer is 42.")
	}
}
