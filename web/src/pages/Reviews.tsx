import { useEffect, useRef, useState } from 'react';
import { Button, Card, Collapse, Empty, Input, List, Modal, Select, Space, Spin, Table, Tag, Typography, Upload, message } from 'antd';
import type { UploadFile } from 'antd';
import { PictureOutlined, RobotOutlined, UserAddOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { api, DomainRoleUsers, Guide, ImageInput, PrecheckFinding, ReviewDomain, ReviewRole, ReviewScenario, User, precheckStream } from '../api/client';
import { useAuthStore } from '../store/auth';

const { Title } = Typography;
const { TextArea } = Input;

export default function Reviews() {
  const [guides, setGuides] = useState<Guide[]>([]);
  const [loading, setLoading] = useState(false);
  const [active, setActive] = useState<Guide | null>(null);
  const [note, setNote] = useState('');
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<ReviewRole[]>([]);
  const [domains, setDomains] = useState<ReviewDomain[]>([]);
  const [scenarios, setScenarios] = useState<ReviewScenario[]>([]);
  const [domainRoleUsers, setDomainRoleUsers] = useState<DomainRoleUsers[]>([]);
  const [domainId, setDomainId] = useState('');
  const [scenarioId, setScenarioId] = useState('');
  const [reviewers, setReviewers] = useState<string[]>([]);
  const [files, setFiles] = useState<UploadFile[]>([]);
  const [prechecking, setPrechecking] = useState(false);
  const [summary, setSummary] = useState('');
  const [findings, setFindings] = useState<PrecheckFinding[]>([]);
  const [streamText, setStreamText] = useState('');
  const [streamDone, setStreamDone] = useState(false);
  const canWrite = useAuthStore((s) => s.user?.role !== 'readonly');
  const streamRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (streamRef.current) streamRef.current.scrollTop = streamRef.current.scrollHeight;
  }, [streamText]);

  const load = () => {
    setLoading(true);
    api.listGuides().then((g) => setGuides(g.filter((x) => x.status === 'reviewing' || x.status === 'draft'))).catch((e) => message.error(e.message)).finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
    Promise.all([
      api.listUsers(),
      api.listReviewRoles(),
      api.listReviewDomains(),
      api.listReviewScenarios(),
    ]).then(([u, r, d, s]) => {
      setUsers(u);
      setRoles(r);
      setDomains(d);
      setScenarios(s);
      setDomainId(d[0]?.id || '');
      setScenarioId(s[0]?.id || '');
      if (d[0]?.id) api.listDomainRoleUsers(d[0].id).then(setDomainRoleUsers).catch(() => {});
    }).catch(() => {});
  }, []);

  const roleMeta = Object.fromEntries(roles.map((r) => [r.key, r]));
  const roleName = (key: string) => roleMeta[key]?.name || (key === 'admin' ? '管理员' : key === 'readonly' ? '只读' : key);
  const usersById = Object.fromEntries(users.map((u) => [u.id, u]));
  const reviewerOptions = users
    .filter((u) => u.role !== 'readonly')
    .map((u) => ({ value: u.id, label: `${u.username} · ${roleName(u.role)}` }));

  const applyDefaults = async (nextDomainId = domainId, nextScenarioId = scenarioId) => {
    if (!nextDomainId || !nextScenarioId) return;
    const roleUsers = nextDomainId === domainId && domainRoleUsers.length > 0
      ? domainRoleUsers
      : await api.listDomainRoleUsers(nextDomainId);
    setDomainRoleUsers(roleUsers);
    const scenario = scenarios.find((s) => s.id === nextScenarioId);
    const requiredRoles = scenario?.roleKeys || [];
    const names = roleUsers
      .filter((item) => requiredRoles.includes(item.roleKey))
      .flatMap((item) => item.userIds)
      .filter((id) => !!usersById[id])
      .filter(Boolean) as string[];
    setReviewers(Array.from(new Set(names)));
  };

  const openReview = async (g: Guide) => {
    setActive(g);
    setNote('');
    setFiles([]);
    setSummary('');
    setFindings([]);
    setStreamText('');
    setStreamDone(false);
    const reviews = await api.listReviews(g.id);
    const existing = reviews.map((r) => r.reviewerUserId || users.find((u) => u.username === r.reviewer)?.id || '').filter(Boolean);
    if (existing.length > 0) {
      setReviewers(existing);
    } else {
      await applyDefaults();
    }
  };

  const ensureReviews = async () => {
    if (!active) return [];
    const existing = await api.listReviews(active.id);
    const existingIds = new Set(existing.map((r) => r.reviewerUserId || users.find((u) => u.username === r.reviewer)?.id || ''));
    for (const userId of reviewers) {
      if (!existingIds.has(userId)) {
        await api.createReview(active.id, userId, usersById[userId]?.username);
      }
    }
    const latest = await api.listReviews(active.id);
    return latest.length > 0 ? latest : [await api.createReview(active.id, reviewers[0] || '', usersById[reviewers[0]]?.username || '评审人')];
  };

  const toImages = async (): Promise<ImageInput[]> => {
    const images = await Promise.all(files.map((f) => new Promise<ImageInput | null>((resolve) => {
      const raw = f.originFileObj;
      if (!raw) return resolve(null);
      const reader = new FileReader();
      reader.onload = () => resolve({ dataUrl: String(reader.result), mimeType: raw.type });
      reader.onerror = () => resolve(null);
      reader.readAsDataURL(raw);
    })));
    return images.filter(Boolean) as ImageInput[];
  };

  const runPrecheck = async () => {
    if (!active) return;
    setPrechecking(true);
    setSummary('');
    setFindings([]);
    setStreamText('');
    setStreamDone(false);
    const images = await toImages();
    try {
      await precheckStream(
        active.content,
        images,
        (delta) => setStreamText((prev) => prev + delta),
        (result) => {
          setSummary(result.summary || 'AI 预审完成');
          setFindings(result.findings || []);
        },
        () => {
          setStreamDone(true);
          setPrechecking(false);
          message.success('AI 预审完成');
        },
        (msg) => {
          setStreamDone(true);
          setPrechecking(false);
          message.error(msg);
        }
      );
    } catch (e: any) {
      setPrechecking(false);
      message.error(e.message);
    }
  };

  const decide = async (status: string) => {
    if (!active) return;
    const reviews = await ensureReviews();
    const rv = reviews[0];
    if (!rv) return message.error('无评审实例');
    try {
      await api.decideReview(rv.id, status, note);
      message.success(status === 'approved' ? '已通过' : '已驳回');
      setActive(null);
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  return (
    <div>
      <Title level={3} style={{ color: 'var(--text-primary)' }}>评审工作台</Title>
      <Card style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={guides}
          pagination={false}
          locale={{ emptyText: <Empty description="暂无待评审的材料" /> }}
          columns={[
            { title: '标题', dataIndex: 'title' },
            { title: '状态', dataIndex: 'status', width: 110, render: (s) => <Tag color={s === 'reviewing' ? 'processing' : 'default'}>{s === 'reviewing' ? '评审中' : '草稿'}</Tag> },
            { title: '风险', dataIndex: 'riskLevel', width: 90 },
            { title: '操作', width: 100, render: (_, r) => <Button size="small" type={canWrite ? 'primary' : 'default'} onClick={() => openReview(r)}>{canWrite ? '评审' : '查看'}</Button> },
          ]}
        />
      </Card>

      <Modal
        title={active?.title}
        open={!!active}
        onCancel={() => setActive(null)}
        width={980}
        footer={canWrite ? [
          <Button key="reject" danger onClick={() => decide('rejected')}>驳回</Button>,
          <Button key="approve" type="primary" onClick={() => decide('approved')}>通过</Button>,
        ] : null}
      >
        <Space direction="vertical" style={{ width: '100%', marginBottom: 12 }} size="middle">
          {canWrite && (
            <>
              <Select
                mode="multiple"
                placeholder="选择评审人"
                value={reviewers}
                onChange={setReviewers}
                options={reviewerOptions}
                suffixIcon={<UserAddOutlined />}
              />
              <Space wrap>
                <Select
                  style={{ minWidth: 180 }}
                  placeholder="选择领域"
                  value={domainId || undefined}
                  options={domains.map((d) => ({ value: d.id, label: d.name }))}
                  onChange={async (value) => {
                    setDomainId(value);
                    await applyDefaults(value, scenarioId);
                  }}
                />
                <Select
                  style={{ minWidth: 180 }}
                  placeholder="选择审批场景"
                  value={scenarioId || undefined}
                  options={scenarios.map((s) => ({ value: s.id, label: s.name }))}
                  onChange={async (value) => {
                    setScenarioId(value);
                    await applyDefaults(domainId, value);
                  }}
                />
                <Button icon={<UserAddOutlined />} onClick={() => applyDefaults()}>
                  带出默认评审人
                </Button>
              </Space>
              <Space wrap>
                <Upload
                  accept="image/*"
                  fileList={files}
                  beforeUpload={() => false}
                  onChange={({ fileList }) => setFiles(fileList)}
                  multiple
                >
                  <Button icon={<PictureOutlined />}>上传评审图片</Button>
                </Upload>
                <Button icon={<RobotOutlined />} loading={prechecking} onClick={runPrecheck}>
                  Hermes AI 预审
                </Button>
              </Space>
            </>
          )}
          {(prechecking || streamText) && findings.length === 0 && (
            <Card
              size="small"
              title={
                <Space>
                  <RobotOutlined style={{ color: 'var(--color-primary, #1677ff)' }} />
                  <span>{prechecking ? 'Hermes AI 正在预审…' : 'AI 原始输出'}</span>
                  {prechecking && <Spin size="small" />}
                </Space>
              }
              style={{ borderColor: 'var(--border-color)' }}
            >
              <div
                ref={streamRef}
                style={{
                  maxHeight: 200,
                  overflow: 'auto',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  fontFamily: 'var(--font-mono, monospace)',
                  fontSize: 12,
                  lineHeight: 1.7,
                  color: 'var(--text-secondary)',
                }}
              >
                {streamText || '正在连接 Hermes Agent，结合已沉淀规则进行流式预审…'}
                {prechecking && <span className="cb-blink">▋</span>}
              </div>
            </Card>
          )}
          {(summary || findings.length > 0) && (
            <Card size="small" title={`AI 预审结果 · 共 ${findings.length} 项`} style={{ borderColor: 'var(--border-color)' }}>
              {summary && <Typography.Paragraph style={{ marginBottom: 8 }}>{summary}</Typography.Paragraph>}
              <List
                size="small"
                dataSource={findings}
                renderItem={(f) => (
                  <List.Item>
                    <Space direction="vertical" size={2}>
                      <Space>
                        <Tag color={f.severity === 'critical' ? 'red' : f.severity === 'warning' ? 'gold' : 'blue'}>{f.severity}</Tag>
                        <Tag>{f.category}</Tag>
                      </Space>
                      <span>{f.problem}</span>
                      <Typography.Text type="secondary">{f.suggestion}</Typography.Text>
                    </Space>
                  </List.Item>
                )}
              />
              {streamText && streamDone && (
                <Collapse
                  ghost
                  size="small"
                  style={{ marginTop: 8 }}
                  items={[{ key: 'raw', label: '查看 AI 原始输出', children: (
                    <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontSize: 12, color: 'var(--text-secondary)', margin: 0 }}>{streamText}</pre>
                  ) }]}
                />
              )}
            </Card>
          )}
        </Space>
        <div className="cb-markdown" style={{ maxHeight: 420, overflow: 'auto', padding: 8, border: '1px solid var(--border-light)', borderRadius: 8 }}>
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{active?.content || ''}</ReactMarkdown>
        </div>
        {canWrite && (
          <Space direction="vertical" style={{ width: '100%', marginTop: 12 }}>
            <TextArea placeholder="评审意见 / 决定说明" rows={3} value={note} onChange={(e) => setNote(e.target.value)} />
          </Space>
        )}
      </Modal>
    </div>
  );
}
