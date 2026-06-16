import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Select, Space, Tag, Typography, message } from 'antd';
import { ApiOutlined, RobotOutlined } from '@ant-design/icons';
import { AgentSettings, AgentType, api } from '../api/client';

const { Title, Paragraph, Text } = Typography;

export default function Settings() {
  const [form] = Form.useForm<AgentSettings>();
  const [types, setTypes] = useState<AgentType[]>([]);
  const [checking, setChecking] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const [settings, agentTypes] = await Promise.all([api.getAgentSettings(), api.getAgentTypes()]);
        form.setFieldsValue(settings);
        setTypes(agentTypes);
      } catch (e: any) {
        message.error(e.message);
      }
    })();
  }, [form]);

  const save = async () => {
    setSaving(true);
    try {
      const values = await form.validateFields();
      const saved = await api.updateAgentSettings(values);
      form.setFieldsValue(saved);
      message.success('模型配置已保存');
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setSaving(false);
    }
  };

  const check = async () => {
    setChecking(true);
    try {
      const res = await api.checkAgentHealth();
      if (res.healthy) message.success(res.message || 'Agent 配置可用');
      else message.error(res.error || 'Agent 配置不可用');
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setChecking(false);
    }
  };

  return (
    <div>
      <Title level={3} style={{ color: 'var(--text-primary)' }}>设置</Title>
      <Paragraph style={{ color: 'var(--text-secondary)' }}>
        配置智能评审使用的 Agent 类型、模型与推理服务地址。保存后，新的生成和预审请求会使用最新配置。
      </Paragraph>

      <Card
        title={<Space><RobotOutlined />Agent 与模型</Space>}
        extra={<Button icon={<ApiOutlined />} loading={checking} onClick={check}>连通性检查</Button>}
        style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}
      >
        <Form form={form} layout="vertical">
          <Space size="large" wrap style={{ width: '100%' }}>
            <Form.Item label="Agent 类型" name="provider" rules={[{ required: true }]} style={{ minWidth: 220 }}>
              <Select
                options={types.map((t) => ({ value: t.type, label: t.name }))}
                optionRender={(option) => {
                  const t = types.find((x) => x.type === option.value);
                  return (
                    <Space direction="vertical" size={0}>
                      <Text>{t?.name || option.label}</Text>
                      {t?.description && <Text type="secondary" style={{ fontSize: 12 }}>{t.description}</Text>}
                    </Space>
                  );
                }}
              />
            </Form.Item>
            <Form.Item label="模型" name="model" rules={[{ required: true, message: '请输入模型 ID' }]} style={{ minWidth: 260 }}>
              <Input placeholder="如 hermes-3 / gpt-4.1 / qwen3" />
            </Form.Item>
            <Form.Item label="超时（秒）" name="timeoutSeconds" style={{ minWidth: 140 }}>
              <InputNumber min={10} max={600} style={{ width: '100%' }} />
            </Form.Item>
          </Space>

          <Space size="large" wrap style={{ width: '100%' }}>
            <Form.Item label="API Base URL" name="baseUrl" style={{ minWidth: 360 }}>
              <Input placeholder="https://your-agent-or-llm-gateway/v1" />
            </Form.Item>
            <Form.Item label="API Key" name="apiKey" style={{ minWidth: 280 }}>
              <Input.Password placeholder="留空或保持掩码则不修改" />
            </Form.Item>
          </Space>

          <Form.Item label="Embedding 模型" name="embeddingModel">
            <Input placeholder="可选；留空则使用应用层降级召回" />
          </Form.Item>

          <Form.Item label="系统提示词" name="systemPrompt">
            <Input.TextArea rows={5} placeholder="定义评审专家角色、审查边界、输出风格与风险偏好" />
          </Form.Item>

          <Space wrap>
            <Button type="primary" loading={saving} onClick={save}>保存模型配置</Button>
            <Tag color="blue">支持 Mock / OpenAI 兼容 / Hermes Agent</Tag>
          </Space>
        </Form>
      </Card>
    </div>
  );
}
