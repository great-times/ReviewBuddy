package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OpenAICompatAdapter 对接任何 OpenAI 兼容 /chat/completions 端点。
// Hermes（Nous Research）通过此适配器接入：把 base_url/api_key/model 指向 Hermes 服务即可。
type OpenAICompatAdapter struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAICompatAdapter(baseURL, apiKey, model string, timeout time.Duration) *OpenAICompatAdapter {
	return &OpenAICompatAdapter{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{Timeout: timeout},
	}
}

func (a *OpenAICompatAdapter) Name() string { return "openai_compat" }

type chatReq struct {
	Model       string  `json:"model"`
	Messages    []any   `json:"messages"`
	Temperature float64 `json:"temperature"`
	Stream      bool    `json:"stream"`
}

type chatResp struct {
	Choices []struct {
		Message Message `json:"message"`
		Delta   Message `json:"delta"`
	} `json:"choices"`
}

func (a *OpenAICompatAdapter) Complete(ctx context.Context, req *CompletionRequest) (string, error) {
	body, _ := json.Marshal(chatReq{Model: a.model, Messages: openAIMessages(req), Temperature: req.Temperature})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	a.setHeaders(httpReq)
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("agent http %d: %s", resp.StatusCode, buf.String())
	}
	var out chatResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("agent returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}

func (a *OpenAICompatAdapter) CompleteStream(ctx context.Context, req *CompletionRequest, onChunk func(Chunk)) error {
	body, _ := json.Marshal(chatReq{Model: a.model, Messages: openAIMessages(req), Temperature: req.Temperature, Stream: true})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return err
	}
	a.setHeaders(httpReq)
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return fmt.Errorf("agent http %d: %s", resp.StatusCode, buf.String())
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var out chatResp
		if err := json.Unmarshal([]byte(data), &out); err != nil {
			continue
		}
		if len(out.Choices) > 0 && out.Choices[0].Delta.Content != "" {
			onChunk(Chunk{Delta: out.Choices[0].Delta.Content})
		}
	}
	onChunk(Chunk{Done: true})
	return scanner.Err()
}

func openAIMessages(req *CompletionRequest) []any {
	out := make([]any, 0, len(req.Messages))
	for i, m := range req.Messages {
		if len(req.Images) == 0 || i != len(req.Messages)-1 || m.Role != "user" {
			out = append(out, m)
			continue
		}
		parts := []map[string]any{{"type": "text", "text": m.Content}}
		for _, img := range req.Images {
			url := img.URL
			if url == "" {
				url = img.DataURL
			}
			if url == "" {
				continue
			}
			parts = append(parts, map[string]any{
				"type":      "image_url",
				"image_url": map[string]string{"url": url},
			})
		}
		out = append(out, map[string]any{"role": m.Role, "content": parts})
	}
	return out
}

func (a *OpenAICompatAdapter) setHeaders(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		r.Header.Set("Authorization", "Bearer "+a.apiKey)
	}
}
