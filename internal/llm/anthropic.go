package llm

import (
	"encoding/json"
	"net/http"
)

// AnthropicDialect implements Anthropic's native /v1/messages format.
type AnthropicDialect struct{ BaseURL string }

func (d AnthropicDialect) Name() string { return "anthropic" }

func (d AnthropicDialect) EncodeRequest(req GenerateRequest, apiKey string) (*HTTPRequest, error) {
	body := map[string]any{
		"model":      req.Model,
		"messages":   req.Messages,
		"max_tokens": 4096,
		"stream":     req.Stream,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := d.BaseURL
	if url == "" {
		url = "https://api.anthropic.com/v1"
	}
	return &HTTPRequest{
		URL:    url + "/messages",
		Method: http.MethodPost,
		Headers: map[string]string{
			"x-api-key":         apiKey,
			"anthropic-version": "2023-06-01",
			"Content-Type":      "application/json",
		},
		Body: data,
	}, nil
}

func (d AnthropicDialect) DecodeStream(chunk []byte) ([]StreamEvent, error) {
	// TODO(E2.2): parse Anthropic SSE events (content_block_delta, message_delta, ...).
	return []StreamEvent{{Type: "done"}}, nil
}
