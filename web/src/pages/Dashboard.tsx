import { useEffect, useState } from 'react';
import { Card, Col, Row, Statistic, Typography } from 'antd';
import { FileTextOutlined, EditOutlined, BulbOutlined, AuditOutlined } from '@ant-design/icons';
import { api } from '../api/client';

const { Title, Paragraph } = Typography;

export default function Dashboard() {
  const [templates, setTemplates] = useState(0);
  const [guides, setGuides] = useState(0);
  const [issues, setIssues] = useState(0);
  const [rules, setRules] = useState(0);

  useEffect(() => {
    api.listTemplates().then((t) => setTemplates(t.length)).catch(() => {});
    api.listGuides().then((g) => setGuides(g.length)).catch(() => {});
    api.metrics().then((m) => {
      setIssues(m.issueCount);
      setRules(m.ruleCount);
    }).catch(() => {});
  }, []);

  const cardStyle = { background: 'var(--bg-container)', borderColor: 'var(--border-color)' };

  return (
    <div>
      <Title level={3} style={{ color: 'var(--text-primary)' }}>概览</Title>
      <Paragraph style={{ color: 'var(--text-secondary)' }}>
        ReviewBuddy 是面向多评审域的智能评审助手，覆盖模板库管理、评审材料 AI 辅助评审、人工意见沉淀和模板反哺。
      </Paragraph>
      <Row gutter={16}>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="模板数" value={templates} prefix={<FileTextOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="评审材料数" value={guides} prefix={<EditOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="沉淀问题" value={issues} prefix={<AuditOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="审查规则" value={rules} prefix={<BulbOutlined />} valueStyle={{ color: 'var(--color-primary)' }} />
          </Card>
        </Col>
      </Row>
      <Card style={{ ...cardStyle, marginTop: 16 }} title="评审闭环">
        <Paragraph style={{ color: 'var(--text-secondary)' }}>
          场景模板 → Hermes AI 预审 → 人工评审意见 → 规则沉淀 → 反向更新模板。
          当前采用 rule 作为结构化规则形态，后续可把高频规则整理为 Hermes Agent skill。
        </Paragraph>
      </Card>
    </div>
  );
}
