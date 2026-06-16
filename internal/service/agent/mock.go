package agent

import (
	"context"
	"strings"
	"time"
)

// MockAdapter 在未配置真实 Agent 时使用，便于本地开发与演示。
type MockAdapter struct{}

func NewMockAdapter() *MockAdapter { return &MockAdapter{} }

func (m *MockAdapter) Name() string { return "mock" }

func (m *MockAdapter) Complete(ctx context.Context, req *CompletionRequest) (string, error) {
	return m.canned(req), nil
}

func (m *MockAdapter) CompleteStream(ctx context.Context, req *CompletionRequest, onChunk func(Chunk)) error {
	text := m.canned(req)
	for _, line := range strings.SplitAfter(text, "\n") {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		onChunk(Chunk{Delta: line})
		time.Sleep(20 * time.Millisecond)
	}
	onChunk(Chunk{Done: true})
	return nil
}

func (m *MockAdapter) canned(req *CompletionRequest) string {
	prompt := ""
	if len(req.Messages) > 0 {
		prompt = req.Messages[len(req.Messages)-1].Content
	}
	if len(req.Images) > 0 {
		prompt += "\n\n（已收到图片附件，真实 Hermes Agent 会结合图片内容评审。）"
	}
	_ = prompt
	return strings.TrimSpace(`
# 评审材料（示例 · Mock 生成）

> 当前为 Mock 适配器输出。配置 Agent 类型、API 地址与模型后即为真实生成。

## 一、背景与目标
说明本次评审材料的背景、目标与适用范围。

## 二、核心内容
- [ ] 目标清晰
- [ ] 方案完整
- [ ] 依赖明确

## 三、风险与约束
1. 风险一……
2. 约束一……

## 四、评审关注点
1. 关注点一……
2. 关注点二……

## 五、结论与后续动作
- [ ] 结论明确
- [ ] 待办责任人清晰
`) + "\n"
}
