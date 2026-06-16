import { useEffect, useState } from 'react';
import { Button, Card, Col, Input, Row, Select, Space, Typography, message } from 'antd';
import { ThunderboltOutlined, SaveOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate } from 'react-router-dom';
import { api, generateGuide, Template, PrecheckFinding } from '../api/client';

const { Title } = Typography;
const { TextArea } = Input;

export default function GuideGenerate() {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [templateId, setTemplateId] = useState<string>();
  const [title, setTitle] = useState('');
  const [changeType, setChangeType] = useState('');
  const [context, setContext] = useState('');
  const [content, setContent] = useState('');
  const [generating, setGenerating] = useState(false);
  const [findings, setFindings] = useState<PrecheckFinding[] | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    api.listTemplates().then(setTemplates).catch(() => {});
  }, []);

  const onGenerate = async () => {
    if (!title) return message.warning('请填写标题');
    setContent('');
    setFindings(null);
    setGenerating(true);
    try {
      await generateGuide(
        { title, templateId, changeType, context },
        (chunk) => setContent((prev) => prev + chunk),
        () => setGenerating(false),
        (err) => { message.error(err); setGenerating(false); }
      );
    } catch (e: any) {
      message.error(e.message);
      setGenerating(false);
    }
  };

  const onSave = async () => {
    if (!title || !content) return message.warning('请先生成内容');
    try {
      await api.createGuide({ title, templateId, content, riskLevel: 'medium' });
      message.success('评审材料已保存为草稿');
      navigate('/guides');
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const onPrecheck = async () => {
    if (!content) return message.warning('请先生成内容');
    try {
      const res = await api.precheck(content);
      setFindings(res.findings || []);
      if (!res.findings?.length) message.success(res.summary || '未发现明显问题');
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const sevColor: Record<string, string> = { critical: 'var(--color-warning)', warning: '#fa8c16', info: 'var(--text-secondary)' };

  return (
    <div>
      <Title level={3} style={{ color: 'var(--text-primary)' }}>AI 生成评审材料</Title>
      <Row gutter={16}>
        <Col span={9}>
          <Card title="生成参数" style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}>
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Input placeholder="评审材料标题" value={title} onChange={(e) => setTitle(e.target.value)} />
              <Select
                placeholder="选择模板（可选）"
                style={{ width: '100%' }}
                allowClear
                value={templateId}
                onChange={setTemplateId}
                options={templates.map((t) => ({ label: t.name, value: t.id }))}
              />
              <Input placeholder="材料类型，如：架构方案 / 测试方案 / 安全评审" value={changeType} onChange={(e) => setChangeType(e.target.value)} />
              <TextArea placeholder="补充上下文：评审对象、目标、范围、约束……" rows={5} value={context} onChange={(e) => setContext(e.target.value)} />
              <Space>
                <Button type="primary" icon={<ThunderboltOutlined />} loading={generating} onClick={onGenerate}>
                  生成
                </Button>
                <Button icon={<SafetyCertificateOutlined />} onClick={onPrecheck}>AI 预审</Button>
                <Button icon={<SaveOutlined />} onClick={onSave}>保存草稿</Button>
              </Space>
            </Space>
          </Card>
          {findings && (
            <Card title="AI 预审结果" style={{ marginTop: 16, background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}>
              {findings.length === 0 ? (
                <span style={{ color: 'var(--text-secondary)' }}>未发现明显问题</span>
              ) : (
                findings.map((f, i) => (
                  <div key={i} style={{ marginBottom: 10, paddingBottom: 10, borderBottom: '1px solid var(--border-light)' }}>
                    <strong style={{ color: sevColor[f.severity] || 'var(--text-primary)' }}>[{f.severity}] {f.category}</strong>
                    <div style={{ color: 'var(--text-primary)' }}>{f.problem}</div>
                    <div style={{ color: 'var(--text-secondary)', fontSize: 13 }}>建议：{f.suggestion}</div>
                  </div>
                ))
              )}
            </Card>
          )}
        </Col>
        <Col span={15}>
          <Card title="生成内容（实时）" style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)', minHeight: 480 }}>
            <div className="cb-markdown">
              {content ? <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown> : <span style={{ color: 'var(--text-secondary)' }}>填写左侧参数后点击「生成」。</span>}
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  );
}
