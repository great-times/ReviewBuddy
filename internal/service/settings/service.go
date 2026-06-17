package settings

import (
	"context"
	"strings"
	"sync"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
	"reviewbuddy/pkg/config"
)

const keyAgent = "agent_settings"

type Service struct {
	repo  *repo.SettingsRepo
	mu    sync.RWMutex
	agent model.AgentSettings
}

func NewService(r *repo.SettingsRepo, cfg config.AgentConfig) *Service {
	s := &Service{
		repo: r,
		agent: model.AgentSettings{
			Provider:       cfg.Provider,
			BaseURL:        cfg.BaseURL,
			APIKey:         cfg.APIKey,
			Model:          cfg.Model,
			EmbeddingModel: cfg.EmbeddingModel,
			TimeoutSeconds: cfg.TimeoutSeconds,
			SystemPrompt:   cfg.SystemPrompt,
		},
	}
	var saved model.AgentSettings
	if ok, err := r.Get(keyAgent, &saved); err == nil && ok {
		if saved.APIKey == "" {
			saved.APIKey = s.agent.APIKey
		}
		s.agent = saved
	}
	if s.agent.Provider == "" {
		s.agent.Provider = "mock"
	}
	if s.agent.Model == "" {
		s.agent.Model = "hermes-3"
	}
	if s.agent.TimeoutSeconds <= 0 {
		s.agent.TimeoutSeconds = 120
	}
	return s
}

func (s *Service) AgentSettings() model.AgentSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := s.agent
	out.APIKey = maskToken(out.APIKey)
	return out
}

func (s *Service) AgentConfig() config.AgentConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return config.AgentConfig{
		Provider:       s.agent.Provider,
		BaseURL:        s.agent.BaseURL,
		APIKey:         s.agent.APIKey,
		Model:          s.agent.Model,
		EmbeddingModel: s.agent.EmbeddingModel,
		TimeoutSeconds: s.agent.TimeoutSeconds,
		SystemPrompt:   s.agent.SystemPrompt,
	}
}

func (s *Service) UpdateAgentSettings(_ context.Context, in model.AgentSettings) (model.AgentSettings, error) {
	s.mu.Lock()
	if in.Provider == "" {
		in.Provider = s.agent.Provider
	}
	if in.Model == "" {
		in.Model = s.agent.Model
	}
	if in.TimeoutSeconds <= 0 {
		in.TimeoutSeconds = 120
	}
	if in.APIKey == "" || isMasked(in.APIKey) {
		in.APIKey = s.agent.APIKey
	}
	s.agent = in
	out := s.agent
	s.mu.Unlock()

	if err := s.repo.Put(keyAgent, out); err != nil {
		return model.AgentSettings{}, err
	}
	out.APIKey = maskToken(out.APIKey)
	return out, nil
}

func AgentTypes() []model.AgentType {
	return []model.AgentType{
		{Type: "mock", Name: "Mock Agent", Description: "本地演示与离线开发使用"},
		{Type: "openai_compat", Name: "OpenAI 兼容", Description: "兼容 /v1/chat/completions 的模型网关"},
		{Type: "hermes", Name: "Hermes Agent", Description: "Hermes 或兼容 Hermes 的私有化 Agent 服务"},
	}
}

func maskToken(token string) string {
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}

func isMasked(token string) bool {
	return strings.Contains(token, "*")
}
