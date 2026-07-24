package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// OpenAIDialect implements the OpenAI-compatible /v1/chat/completions format.
type OpenAIDialect struct{ BaseURL string }

func (d OpenAIDialect) Name() string { return "openai_compat" }

func (d OpenAIDialect) EncodeRequest(req GenerateRequest, apiKey string) (*HTTPRequest, error) {
	body := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   req.Stream,
	}
	if len(req.Tools) > 0 {
		body["tools"] = req.Tools
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := d.BaseURL
	if url == "" {
		url = "https://api.openai.com/v1"
	}
	return &HTTPRequest{
		URL:    url + "/chat/completions",
		Method: http.MethodPost,
		Headers: map[string]string{
			"Authorization": "Bearer " + apiKey,
			"Content-Type":  "application/json",
		},
		Body: data,
	}, nil
}

func (d OpenAIDialect) DecodeStream(chunk []byte) ([]StreamEvent, error) {
	// TODO(E2.1): parse SSE "data: {...}" lines into deltas, tool calls, and usage.
	line := bytes.TrimSpace(chunk)
	if len(line) == 0 || bytes.Equal(line, []byte("data: [DONE]")) {
		return []StreamEvent{{Type: "done"}}, nil
	}
	return []StreamEvent{{Type: "token", Delta: string(line)}}, nil
}

// HTTPProvider is a Provider that drives any Dialect over HTTP.
type HTTPProvider struct {
	IDStr  string
	Dial   Dialect
	APIKey string
	Client *http.Client
}

func (p *HTTPProvider) ID() string       { return p.IDStr }
func (p *HTTPProvider) Dialect() Dialect { return p.Dial }

func (p *HTTPProvider) client() *http.Client {
	if p.Client != nil {
		return p.Client
	}
	return http.DefaultClient
}

func (p *HTTPProvider) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	req.Stream = false
	hr, err := p.Dial.EncodeRequest(req, p.APIKey)
	if err != nil {
		return GenerateResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, hr.Method, hr.URL, bytes.NewReader(hr.Body))
	if err != nil {
		return GenerateResponse{}, err
	}
	for k, v := range hr.Headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := p.client().Do(httpReq)
	if err != nil {
		return GenerateResponse{}, err
	}
	defer resp.Body.Close()
	// TODO(E2.1): decode the full response body via the dialect into GenerateResponse.
	return GenerateResponse{}, nil
}

func (p *HTTPProvider) Stream(ctx context.Context, req GenerateRequest) (<-chan StreamEvent, error) {
	req.Stream = true
	out := make(chan StreamEvent)
	hr, err := p.Dial.EncodeRequest(req, p.APIKey)
	if err != nil {
		return nil, err
	}
	go func() {
		defer close(out)
		httpReq, err := http.NewRequestWithContext(ctx, hr.Method, hr.URL, bytes.NewReader(hr.Body))
		if err != nil {
			out <- StreamEvent{Type: "error", Err: err}
			return
		}
		for k, v := range hr.Headers {
			httpReq.Header.Set(k, v)
		}
		resp, err := p.client().Do(httpReq)
		if err != nil {
			out <- StreamEvent{Type: "error", Err: err}
			return
		}
		defer resp.Body.Close()
		// TODO(E2.1): read SSE lines from resp.Body and decode via the dialect into out.
		out <- StreamEvent{Type: "done"}
	}()
	return out, nil
}
