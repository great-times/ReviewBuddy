# ReviewBuddy · 智能评审助手

> **让每一次评审，都能沉淀成下一次更好的模板与规则。**
>
> 模板库 · Hermes AI 辅助评审 · 人工评审意见 · 规则沉淀 · 模板反哺。

ReviewBuddy 是一个面向多评审域的智能评审助手。它以场景模板库为入口，对评审材料进行 Hermes Agent AI 辅助评审，再由人工给出最终评审意见；人工意见会沉淀为 AI 评审 rule，并反向更新到模板中。高频、稳定的规则后续可整理成 Hermes Agent skill，用于跨场景复用。

完整设计见 [docs/技术方案.md](docs/技术方案.md)。

## 一个平台，多个评审域

平台核心闭环领域无关。一个**评审域** = 一套模板集 + 一套规则集。新增评审域只需配置模板与规则，**无需改动平台**。

| 评审域 | 方案产物 | 典型规则集 |
|--------|----------|------------|
| 材料评审 | 评审材料 | 风险应对完整性、检查项、风险评估 |
| 需求评审 | 需求方案 | 需求清晰度、验收标准、边界与依赖 |
| 架构评审 | 架构设计方案 | 可扩展性、容量、单点、安全合规 |
| 技术方案评审 | 技术方案文档 | 可行性、成本、风险、替代方案对比 |

## 核心闭环

```
方案生成（模板 + AI + RAG 召回历史经验）
      ↓
评审工作台（页面配置评审人 + Hermes AI 预审，结合已沉淀规则）
      ↓
规则沉淀（人工评审意见 → rule 候选 → 可升级为 Hermes skill）
      ↓
模板进化（高频问题提炼为规则 → 反哺模板迭代，质量曲线可量化）
      ↘──────────────── 回流到生成与评审 ───────────────↗
```

这套闭环不依赖模型重训，纯应用层工程实现；底座 LLM/Agent 可插拔。

## 技术栈

- 后端：Go + Gin，SQLite（`modernc.org/sqlite` 纯 Go 驱动，无需 CGO）
- 前端：React 18 + TypeScript + Vite + Ant Design 5 + Zustand
- LLM/Agent：可插拔适配器。Hermes Agent 通过 OpenAI 兼容端点接入，支持图片多模态消息；未配置时回退 Mock 适配器，便于本地演示。
- 部署：纯 Web 云化部署。

## 快速开始

### 1. 后端

```bash
cp configs/config.yaml.example configs/config.yaml   # 首次：填写 Hermes Agent 端点；不填则用 mock
go run ./cmd/server                                   # 监听 :26405
```

接入 Hermes / 任意 OpenAI 兼容服务，编辑 `configs/config.yaml`：

```yaml
agent:
  provider: openai_compat
  base_url: https://your-hermes-endpoint/v1
  api_key: "***"
  model: hermes-3
```

### 2. 前端

```bash
cd web && npm install && npm run dev   # 监听 :26406，/api 代理到后端
```

打开 http://localhost:26406

## 目录结构

```
cmd/server/            主入口（路由注册、依赖注入）
internal/
  api/                 HTTP 处理器（auth/template/guide/review/knowledge/settings/user）
  service/             业务逻辑
    agent/             LLM/Agent 可插拔适配器（mock / openai_compat）
    template/          模板库与模板管理
    guide/             评审材料生成 + 评审
    knowledge/         规则库 / RAG 召回 / 自学习
    settings/          Agent 与模型运行时配置
    user/              用户与角色管理
  repo/                数据访问层（原生 SQL）
  model/               数据模型
  db/                  SQLite 打开 + 幂等建表
pkg/config/            配置加载
sql-change/            版本化迁移 SQL
web/                   React 前端（主题系统）
docs/技术方案.md        完整设计文档
```

## 已实现（M0–M3 雏形）

- 模板库 + 模板 CRUD + 版本化（Monaco 编辑器），按评审域分类
- AI 流式生成（SSE），生成时 RAG 召回历史经验注入 Prompt
- 评审工作台：页面配置评审人、通过/驳回、状态流转
- AI 预审（结合已沉淀规则，支持图片）
- 规则库 / 自学习：评审问题沉淀、审查规则、质量度量、反向更新模板
- 用户管理：管理员、只读、开发、运维、测试、架构、设计
- 设置页运行时维护 Agent 类型、模型、API Base URL、API Key、Embedding 模型、系统提示词

## 待办

- 内置多评审域的模板集与规则集（需求 / 架构 / 测试 / 安全 / 技术方案）
- embedding 向量召回，替换关键词降级方案
- rule 到 Hermes Agent skill 的自动整理、复盘回流、版本 Diff 可视化、质量度量看板增强
