package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// OpenAIProvider talks to any OpenAI-compatible chat-completions endpoint with
// zero external dependencies - we own the wire format. This single client covers
// OpenAI, OpenRouter, Groq, Mistral, Cerebras, DeepInfra, Together, the Alibaba /
// Qwen compatible-mode, Ollama, vLLM, LM Studio, and Mímir's own gateway (F1).
type OpenAIProvider struct {
	IDStr   string
	BaseURL string // e.g. https://api.openai.com/v1 (trailing slash optional)
	APIKey  string // bearer token; may be empty for local servers
	Model   string // default model when a request leaves Model blank
	Client  *http.Client
}

func (p *OpenAIProvider) ID() string { return p.IDStr }

type oaFunction struct {
	Name        string `json:"name"`
	Arguments   string `json:"arguments"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type oaToolCall struct {
	Index    int        `json:"index,omitempty"` // streaming only: which call a delta belongs to
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Function oaFunction `json:"function"`
}

type oaTool struct {
	Type     string     `json:"type"`
	Function oaFunction `json:"function"`
}

type oaMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content"`
	ToolCalls  []oaToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

type oaChoice struct {
	Message      oaMessage `json:"message"`
	Delta        oaMessage `json:"delta"`
	FinishReason string    `json:"finish_reason"`
}

type oaRequest struct {
	Model    string      `json:"model"`
	Messages []oaMessage `json:"messages"`
	Tools    []oaTool    `json:"tools,omitempty"`
	Stream   bool        `json:"stream"`
}

type oaResponse struct {
	Choices []oaChoice `json:"choices"`
}

func (p *OpenAIProvider) httpClient() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

func (p *OpenAIProvider) baseURL() string { return strings.TrimRight(p.BaseURL, "/") }

func (p *OpenAIProvider) model(m string) string {
	if m != "" {
		return m
	}
	if p.Model != "" {
		return p.Model
	}
	return "default"
}

func toOpenAI(messages []Message) []oaMessage {
	out := make([]oaMessage, len(messages))
	for i, m := range messages {
		om := oaMessage{Role: m.Role, Content: m.Content, ToolCallID: m.ToolCallID}
		for _, tc := range m.ToolCalls {
			om.ToolCalls = append(om.ToolCalls, oaToolCall{
				ID: tc.ID, Type: "function",
				Function: oaFunction{Name: tc.Name, Arguments: tc.Arguments},
			})
		}
		out[i] = om
	}
	return out
}

func toOATools(schemas []ToolSchema) []oaTool {
	if len(schemas) == 0 {
		return nil
	}
	out := make([]oaTool, len(schemas))
	for i, s := range schemas {
		params := s.Parameters
		if params == nil {
			params = map[string]any{"type": "object"}
		}
		out[i] = oaTool{Type: "function", Function: oaFunction{
			Name: s.Name, Description: s.Description, Parameters: params,
		}}
	}
	return out
}

// thinkOpen / thinkClose are the reasoning-model chain-of-thought delimiters. They
// are built by concatenation so the literal tag tokens never appear in this source
// file (which keeps tooling that scans for thinking blocks from mis-parsing it).
const (
	thinkOpen  = "<" + "think>"
	thinkClose = "<" + "/think>"
)

// stripThink removes a reasoning chain-of-thought block (the think delimiters) from
// a model reply so the visible answer (and stored memory) is clean. Reasoning models
// (e.g. Qwen3) wrap their thinking in these tags; the answer follows the closing tag.
func stripThink(s string) string {
	if i := strings.Index(s, thinkOpen); i >= 0 {
		if j := strings.Index(s[i:], thinkClose); j >= 0 {
			return strings.TrimSpace(s[i+j+len(thinkClose):])
		}
		return strings.TrimSpace(s[:i])
	}
	return s
}

// thinkStripper removes reasoning chain-of-thought blocks (the think delimiters)
// from a stream of text deltas - the streaming counterpart of stripThink. Reasoning
// models emit their thinking first; only the answer after the closing tag should
// reach the user. It buffers a few bytes at chunk boundaries so a delimiter split
// across two deltas is still detected.
type thinkStripper struct {
	buf     string
	inThink bool
}

// feed adds a delta and returns the visible text to emit now.
func (t *thinkStripper) feed(d string) string {
	t.buf += d
	var out strings.Builder
	for {
		if t.inThink {
			if i := strings.Index(t.buf, thinkClose); i >= 0 {
				t.buf = t.buf[i+len(thinkClose):]
				t.inThink = false
				continue
			}
			t.buf = tail(t.buf, len(thinkClose)-1)
			return out.String()
		}
		if i := strings.Index(t.buf, thinkOpen); i >= 0 {
			out.WriteString(t.buf[:i])
			t.buf = t.buf[i+len(thinkOpen):]
			t.inThink = true
			continue
		}
		keep := len(thinkOpen) - 1
		if keep > len(t.buf) {
			keep = len(t.buf)
		}
		out.WriteString(t.buf[:len(t.buf)-keep])
		t.buf = t.buf[len(t.buf)-keep:]
		return out.String()
	}
}

// flush returns any remaining visible text at end of stream (suppressing it if the
// stream ended mid-think, matching stripThink's behavior).
func (t *thinkStripper) flush() string {
	if t.inThink {
		t.buf = ""
		return ""
	}
	out := t.buf
	t.buf = ""
	return out
}

func tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func (p *OpenAIProvider) post(ctx context.Context, req GenerateRequest) (*http.Response, error) {
	body, err := json.Marshal(oaRequest{
		Model:    p.model(req.Model),
		Messages: toOpenAI(req.Messages),
		Tools:    toOATools(req.Tools),
		Stream:   req.Stream,
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL()+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	return p.httpClient().Do(httpReq)
}

// Generate returns a complete (non-streamed) reply, including any tool calls.
func (p *OpenAIProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	req.Stream = false
	resp, err := p.post(ctx, req)
	if err != nil {
		return GenerateResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return GenerateResponse{}, fmt.Errorf("provider %s: HTTP %d: %s", p.IDStr, resp.StatusCode, string(b))
	}
	var r oaResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return GenerateResponse{}, err
	}
	if len(r.Choices) == 0 {
		return GenerateResponse{}, fmt.Errorf("provider %s: empty choices", p.IDStr)
	}
	ch := r.Choices[0]
	var calls []ToolCall
	for _, tc := range ch.Message.ToolCalls {
		calls = append(calls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: tc.Function.Arguments})
	}
	return GenerateResponse{Content: stripThink(ch.Message.Content), ToolCalls: calls, FinishReason: ch.FinishReason}, nil
}

// Stream returns a channel of streamed events: text deltas (with reasoning
// think-blocks stripped), ending with a Done event that carries any tool calls the
// model issued (accumulated across the stream's fragmented tool-call deltas).
func (p *OpenAIProvider) Stream(ctx context.Context, req GenerateRequest) (<-chan StreamEvent, error) {
	req.Stream = true
	resp, err := p.post(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("provider %s: HTTP %d: %s", p.IDStr, resp.StatusCode, string(b))
	}
	out := make(chan StreamEvent)
	go func() {
		defer close(out)
		defer resp.Body.Close()
		var ts thinkStripper
		toolAcc := map[int]*ToolCall{}
		finish := func() {
			if vis := ts.flush(); vis != "" {
				out <- StreamEvent{Delta: vis}
			}
			var tcs []ToolCall
			if len(toolAcc) > 0 {
				idxs := make([]int, 0, len(toolAcc))
				for i := range toolAcc {
					idxs = append(idxs, i)
				}
				sort.Ints(idxs)
				for _, i := range idxs {
					tcs = append(tcs, *toolAcc[i])
				}
			}
			out <- StreamEvent{Done: true, ToolCalls: tcs}
		}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "" {
				continue
			}
			if data == "[DONE]" {
				finish()
				return
			}
			var r oaResponse
			if err := json.Unmarshal([]byte(data), &r); err != nil {
				continue
			}
			if len(r.Choices) == 0 {
				continue
			}
			ch := r.Choices[0]
			if d := ch.Delta.Content; d != "" {
				if vis := ts.feed(d); vis != "" {
					out <- StreamEvent{Delta: vis}
				}
			}
			for _, tc := range ch.Delta.ToolCalls {
				cur, ok := toolAcc[tc.Index]
				if !ok {
					cur = &ToolCall{}
					toolAcc[tc.Index] = cur
				}
				if tc.ID != "" {
					cur.ID = tc.ID
				}
				if tc.Function.Name != "" {
					cur.Name = tc.Function.Name
				}
				cur.Arguments += tc.Function.Arguments
			}
			if ch.FinishReason != "" {
				finish()
				return
			}
		}
		if err := scanner.Err(); err != nil {
			out <- StreamEvent{Err: err}
			return
		}
		finish()
	}()
	return out, nil
}
