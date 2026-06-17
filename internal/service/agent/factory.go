package agent

import (
	"context"
	"time"

	"reviewbuddy/pkg/config"
)

// New 按配置创建适配器。provider 未配置或为 mock，或缺少 base_url 时回退到 Mock。
func New(cfg config.AgentConfig) Adapter {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	switch cfg.Provider {
	case "openai_compat", "hermes":
		if cfg.BaseURL == "" {
			return NewMockAdapter()
		}
		return NewOpenAICompatAdapter(cfg.BaseURL, cfg.APIKey, cfg.Model, timeout)
	default:
		return NewMockAdapter()
	}
}

type ConfigProvider interface {
	AgentConfig() config.AgentConfig
}

type DynamicAdapter struct {
	provider ConfigProvider
}

func NewDynamicAdapter(provider ConfigProvider) *DynamicAdapter {
	return &DynamicAdapter{provider: provider}
}

func (a *DynamicAdapter) Name() string {
	return New(a.provider.AgentConfig()).Name()
}

func (a *DynamicAdapter) Complete(ctx context.Context, req *CompletionRequest) (string, error) {
	cfg := a.provider.AgentConfig()
	return New(cfg).Complete(ctx, withSystemPrompt(req, cfg.SystemPrompt))
}

func (a *DynamicAdapter) CompleteStream(ctx context.Context, req *CompletionRequest, onChunk func(Chunk)) error {
	cfg := a.provider.AgentConfig()
	return New(cfg).CompleteStream(ctx, withSystemPrompt(req, cfg.SystemPrompt), onChunk)
}

func withSystemPrompt(req *CompletionRequest, prompt string) *CompletionRequest {
	if prompt == "" || req == nil {
		return req
	}
	next := *req
	next.Messages = append([]Message{{Role: "system", Content: prompt}}, req.Messages...)
	return &next
}
