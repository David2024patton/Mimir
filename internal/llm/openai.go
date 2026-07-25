package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type oaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type oaChoice struct {
	Message      oaMessage `json:"message"`
	Delta        oaMessage `json:"delta"`
	FinishReason string    `json:"finish_reason"`
}

type oaRequest struct {
	Model    string      `json:"model"`
	Messages []oaMessage `json:"messages"`
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
		out[i] = oaMessage{Role: m.Role, Content: m.Content}
	}
	return out
}

func (p *OpenAIProvider) post(ctx context.Context, req GenerateRequest) (*http.Response, error) {
	body, err := json.Marshal(oaRequest{
		Model:    p.model(req.Model),
		Messages: toOpenAI(req.Messages),
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

// Generate returns a complete (non-streamed) reply.
func (p *OpenAIProvider) Generate(ctx context.Context, req GenerateRequest) (string, error) {
	req.Stream = false
	resp, err := p.post(ctx, req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("provider %s: HTTP %d: %s", p.IDStr, resp.StatusCode, string(b))
	}
	var r oaResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if len(r.Choices) == 0 {
		return "", fmt.Errorf("provider %s: empty choices", p.IDStr)
	}
	return r.Choices[0].Message.Content, nil
}

// Stream returns a channel of token deltas, ending with a Done event.
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
				out <- StreamEvent{Done: true}
				return
			}
			var r oaResponse
			if err := json.Unmarshal([]byte(data), &r); err != nil {
				continue
			}
			if len(r.Choices) == 0 {
				continue
			}
			if d := r.Choices[0].Delta.Content; d != "" {
				out <- StreamEvent{Delta: d}
			}
			if r.Choices[0].FinishReason != "" {
				out <- StreamEvent{Done: true}
				return
			}
		}
		if err := scanner.Err(); err != nil {
			out <- StreamEvent{Err: err}
		}
	}()
	return out, nil
}
