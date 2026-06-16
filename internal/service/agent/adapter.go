package agent

import "context"

// Chunk 流式输出片段
type Chunk struct {
	Delta string `json:"delta"`
	Done  bool   `json:"done"`
}

// Message 对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ImageInput struct {
	URL      string `json:"url,omitempty"`
	DataURL  string `json:"dataUrl,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}

// CompletionRequest 推理请求
type CompletionRequest struct {
	Messages    []Message
	Images      []ImageInput
	Temperature float64
}

// Adapter 是 LLM/Agent 的可插拔抽象。Hermes Agent 通过 openai_compat 适配器接入。
type Adapter interface {
	Name() string
	Complete(ctx context.Context, req *CompletionRequest) (string, error)
	CompleteStream(ctx context.Context, req *CompletionRequest, onChunk func(Chunk)) error
}
