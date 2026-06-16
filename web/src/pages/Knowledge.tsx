import { useEffect, useState } from 'react';
import { Button, Card, Col, Form, Input, Modal, Row, Table, Tag, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { http } from '../api/client';
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
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();
  const canWrite = useAuthStore((s) => s.user?.role !== 'readonly');

  const load = () => {
    http.get('/knowledge/issues').then((r) => setIssues(r.data.data)).catch(() => {});
    http.get('/knowledge/rules').then((r) => setRules(r.data.data)).catch(() => {});
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

  const cardStyle = { background: 'var(--bg-container)', borderColor: 'var(--border-color)' };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={3} style={{ color: 'var(--text-primary)' }}>规则沉淀</Title>
        {canWrite && <Button type="primary" icon={<PlusOutlined />} onClick={() => setOpen(true)}>沉淀问题</Button>}
      </div>
      <Paragraph style={{ color: 'var(--text-secondary)' }}>
        人工评审意见先沉淀为 rule，生成与 AI 预审时自动召回；高频规则可再整理为 Hermes Agent skill，并反向更新模板。
      </Paragraph>
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
