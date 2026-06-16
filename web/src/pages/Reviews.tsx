import { useEffect, useState } from 'react';
import { Button, Card, Empty, Input, List, Modal, Select, Space, Table, Tag, Typography, Upload, message } from 'antd';
import type { UploadFile } from 'antd';
import { PictureOutlined, RobotOutlined, UserAddOutlined } from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { api, Guide, ImageInput, PrecheckFinding, User } from '../api/client';
import { useAuthStore } from '../store/auth';

const { Title } = Typography;
const { TextArea } = Input;

export default function Reviews() {
  const [guides, setGuides] = useState<Guide[]>([]);
  const [loading, setLoading] = useState(false);
  const [active, setActive] = useState<Guide | null>(null);
  const [note, setNote] = useState('');
  const [users, setUsers] = useState<User[]>([]);
  const [reviewers, setReviewers] = useState<string[]>([]);
  const [files, setFiles] = useState<UploadFile[]>([]);
  const [prechecking, setPrechecking] = useState(false);
  const [summary, setSummary] = useState('');
  const [findings, setFindings] = useState<PrecheckFinding[]>([]);
  const canWrite = useAuthStore((s) => s.user?.role !== 'readonly');

  const load = () => {
    setLoading(true);
    api.listGuides().then((g) => setGuides(g.filter((x) => x.status === 'reviewing' || x.status === 'draft'))).catch((e) => message.error(e.message)).finally(() => setLoading(false));
  };
  useEffect(() => {
    load();
    api.listUsers().then(setUsers).catch(() => {});
  }, []);

  const openReview = async (g: Guide) => {
    setActive(g);
    setNote('');
    setFiles([]);
    setSummary('');
    setFindings([]);
    const reviews = await api.listReviews(g.id);
    setReviewers(reviews.map((r) => r.reviewer).filter(Boolean));
  };

  const ensureReviews = async () => {
    if (!active) return [];
    const existing = await api.listReviews(active.id);
    const existingNames = new Set(existing.map((r) => r.reviewer));
    for (const name of reviewers) {
      if (!existingNames.has(name)) {
        await api.createReview(active.id, name);
      }
    }
    const latest = await api.listReviews(active.id);
    return latest.length > 0 ? latest : [await api.createReview(active.id, reviewers[0] || '评审人')];
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
    try {
      const result = await api.precheck(active.content, await toImages());
      setSummary(result.summary || 'AI 预审完成');
      setFindings(result.findings || []);
      message.success('AI 预审完成');
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setPrechecking(false);
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
                options={users.filter((u) => u.role !== 'readonly').map((u) => ({ value: u.username, label: `${u.username} · ${u.role}` }))}
                suffixIcon={<UserAddOutlined />}
              />
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
          {(summary || findings.length > 0) && (
            <Card size="small" title="AI 预审结果" style={{ borderColor: 'var(--border-color)' }}>
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
