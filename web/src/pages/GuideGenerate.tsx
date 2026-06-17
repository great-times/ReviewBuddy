import { useEffect, useState } from 'react';
import { Alert, Button, Card, Col, Collapse, Input, List, Row, Select, Space, Spin, Tag, Typography, message } from 'antd';
import { ThunderboltOutlined, SaveOutlined, SafetyCertificateOutlined, RobotOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { useNavigate } from 'react-router-dom';
import { api, generateGuide, Template, PrecheckFinding, precheckStream } from '../api/client';

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
  const [prechecking, setPrechecking] = useState(false);
  const [findings, setFindings] = useState<PrecheckFinding[] | null>(null);
  const [precheckSummary, setPrecheckSummary] = useState('');
  const [precheckRaw, setPrecheckRaw] = useState('');
  const [parseOk, setParseOk] = useState<boolean | undefined>();
  const [aiSteps, setAiSteps] = useState<string[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    api.listTemplates().then(setTemplates).catch(() => {});
  }, []);

  const onGenerate = async () => {
    if (!title) return message.warning('请填写标题');
    setContent('');
    setFindings(null);
    setPrecheckSummary('');
    setPrecheckRaw('');
    setParseOk(undefined);
    setAiSteps([
      templateId ? '已读取所选模板' : '未选择模板，将使用标准结构生成',
      changeType ? '已识别材料类型' : '材料类型为空，将由上下文推断',
      context ? '已加载补充上下文' : '上下文较少，生成结果可能需要人工补充',
      '正在召回已沉淀规则并流式生成',
    ]);
    setGenerating(true);
    try {
      await generateGuide(
        { title, templateId, changeType, context },
        (chunk) => setContent((prev) => prev + chunk),
        () => {
          setGenerating(false);
          setAiSteps((prev) => [...prev, '生成完成，可继续 AI 预审或保存草稿']);
        },
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
    setPrechecking(true);
    setFindings(null);
    setPrecheckSummary('');
    setPrecheckRaw('');
    setParseOk(undefined);
    try {
      await precheckStream(
        content,
        [],
        (chunk) => setPrecheckRaw((prev) => prev + chunk),
        (res) => {
          setPrecheckSummary(res.summary || '');
          setFindings(res.findings || []);
          setParseOk(res.parseOk);
        },
        () => {
          setPrechecking(false);
          message.success('AI 预审完成');
        },
        (msg) => {
          setPrechecking(false);
          message.error(msg);
        }
      );
    } catch (e: any) {
      setPrechecking(false);
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
                <Button icon={<SafetyCertificateOutlined />} loading={prechecking} onClick={onPrecheck}>AI 预审</Button>
                <Button icon={<SaveOutlined />} onClick={onSave}>保存草稿</Button>
              </Space>
            </Space>
          </Card>
          <Card
            title={<Space><RobotOutlined />AI 工作状态 {generating && <Spin size="small" />}</Space>}
            style={{ marginTop: 16, background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              {aiSteps.length === 0 ? (
                <Typography.Text type="secondary">AI 会读取模板、结合上下文和沉淀规则生成评审材料。</Typography.Text>
              ) : (
                aiSteps.map((step, idx) => <Tag key={idx} color={idx === aiSteps.length - 1 && generating ? 'processing' : 'blue'}>{step}</Tag>)
              )}
            </Space>
          </Card>
          {(findings || prechecking || precheckRaw) && (
            <Card title="AI 预审结果" style={{ marginTop: 16, background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}>
              {parseOk === false && <Alert type="warning" showIcon message="AI 返回内容未能完全结构化，已保留原始输出供检查。" style={{ marginBottom: 12 }} />}
              {prechecking && <Spin size="small" />}
              {precheckSummary && <Typography.Paragraph>{precheckSummary}</Typography.Paragraph>}
              {(findings || []).length === 0 ? (
                <span style={{ color: 'var(--text-secondary)' }}>未发现明显问题</span>
              ) : (
                <List
                  size="small"
                  dataSource={findings || []}
                  renderItem={(f, i) => (
                    <List.Item key={i}>
                      <Space direction="vertical" size={2}>
                        <strong style={{ color: sevColor[f.severity] || 'var(--text-primary)' }}>[{f.severity}] {f.category}</strong>
                        <div style={{ color: 'var(--text-primary)' }}>{f.problem}</div>
                        <div style={{ color: 'var(--text-secondary)', fontSize: 13 }}>建议：{f.suggestion}</div>
                      </Space>
                    </List.Item>
                  )}
                />
              )}
              {precheckRaw && (
                <Collapse ghost size="small" items={[{ key: 'raw', label: '查看 AI 原始输出', children: <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>{precheckRaw}</pre> }]} />
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
