import { useEffect, useState } from 'react';
import { Button, Card, Col, Form, Input, List, Modal, Row, Space, Table, Tag, Typography, message } from 'antd';
import { CheckOutlined, PlusOutlined, RobotOutlined } from '@ant-design/icons';
import { api, http, LearningSuggestion } from '../api/client';
import { useAuthStore } from '../store/auth';

const { Title, Paragraph } = Typography;

interface Issue {
  id: string;
  category: string;
  problemDesc: string;
  correctPractice: string;
  changeType: string;
  frequency: number;
}
interface Rule {
  id: string;
  title: string;
  ruleType: string;
  suggestion: string;
  hitCount: number;
  enabled: boolean;
}

export default function Knowledge() {
  const [issues, setIssues] = useState<Issue[]>([]);
  const [rules, setRules] = useState<Rule[]>([]);
  const [suggestions, setSuggestions] = useState<LearningSuggestion[]>([]);
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();
  const canWrite = useAuthStore((s) => s.user?.role !== 'readonly');

  const load = () => {
    http.get('/knowledge/issues').then((r) => setIssues(r.data.data)).catch(() => {});
    http.get('/knowledge/rules').then((r) => setRules(r.data.data)).catch(() => {});
    api.listLearningSuggestions('pending').then(setSuggestions).catch(() => {});
  };
  useEffect(load, []);

  const addIssue = async () => {
    const v = await form.validateFields();
    try {
      await http.post('/knowledge/issues', v);
      message.success('已沉淀到知识库');
      setOpen(false);
      form.resetFields();
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const applySuggestion = async (id: string) => {
    try {
      await api.applyLearningSuggestion(id);
      message.success('已应用 AI 沉淀建议');
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const cardStyle = { background: 'var(--bg-container)', borderColor: 'var(--border-color)' };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={3} style={{ color: 'var(--text-primary)' }}>规则沉淀</Title>
        {canWrite && <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>沉淀问题</Button>}
      </div>
      <Paragraph style={{ color: 'var(--text-secondary)' }}>
        人工评审意见会先由 AI 提炼为规则候选和模板更新建议，确认后再进入规则库并反向更新模板。
      </Paragraph>
      <Card
        title={<Space><RobotOutlined />AI 提炼候选</Space>}
        style={{ ...cardStyle, marginBottom: 16 }}
      >
        <List
          dataSource={suggestions}
          locale={{ emptyText: '暂无待确认的 AI 沉淀候选' }}
          renderItem={(item) => (
            <List.Item
              actions={canWrite ? [
                <Button key="apply" size="small" type="primary" icon={<CheckOutlined />} onClick={() => applySuggestion(item.id)}>应用</Button>,
              ] : []}
            >
              <List.Item.Meta
                title={<Space><Tag color="blue">待确认</Tag><span>{item.summary || 'AI 已提炼评审沉淀候选'}</span></Space>}
                description={(
                  <Space direction="vertical" size={4} style={{ width: '100%' }}>
                    <Typography.Text type="secondary">原始意见：{item.rawNote}</Typography.Text>
                    <div>
                      <Tag>问题 {item.issues?.length || 0}</Tag>
                      <Tag color="green">规则 {item.rules?.length || 0}</Tag>
                      {item.templateSuggestion && <Tag color="purple">模板建议</Tag>}
                    </div>
                    {item.rules?.map((rule, idx) => (
                      <Typography.Text key={idx}>规则：{rule.title} - {rule.suggestion}</Typography.Text>
                    ))}
                    {item.templateSuggestion && (
                      <Typography.Text type="secondary">模板建议：{item.templateSuggestion}</Typography.Text>
                    )}
                  </Space>
                )}
              />
            </List.Item>
          )}
        />
      </Card>
      <Row gutter={16}>
        <Col span={12}>
          <Card title="沉淀的评审问题" style={cardStyle}>
            <Table
              rowKey="id"
              dataSource={issues}
              pagination={false}
              size="small"
              columns={[
                { title: '类别', dataIndex: 'category', width: 90, render: (c) => <Tag>{c}</Tag> },
                { title: '问题', dataIndex: 'problemDesc' },
                { title: '频次', dataIndex: 'frequency', width: 60 },
              ]}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="审查规则" style={cardStyle}>
            <Table
              rowKey="id"
              dataSource={rules}
              pagination={false}
              size="small"
              columns={[
                { title: '规则', dataIndex: 'title' },
                { title: '类型', dataIndex: 'ruleType', width: 100, render: (t) => <Tag color="green">{t}</Tag> },
                { title: '命中', dataIndex: 'hitCount', width: 60 },
              ]}
            />
          </Card>
        </Col>
      </Row>

      <Modal title="沉淀评审问题" open={open} onCancel={() => setOpen(false)} onOk={addIssue} okText="保存" cancelText="取消">
        <Form form={form} layout="vertical">
          <Form.Item name="category" label="类别" rules={[{ required: true }]}>
            <Input placeholder="如：风险应对 / 检查项 / 风险评估" />
          </Form.Item>
          <Form.Item name="changeType" label="关联材料类型">
            <Input placeholder="如：数据库方案" />
          </Form.Item>
          <Form.Item name="problemDesc" label="问题描述" rules={[{ required: true }]}>
            <Input.TextArea rows={2} placeholder="发现了什么问题" />
          </Form.Item>
          <Form.Item name="correctPractice" label="正确做法">
            <Input.TextArea rows={2} placeholder="应该怎么做" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
