import { useEffect, useMemo, useState } from 'react';
import { Card, Col, Empty, List, Row, Space, Statistic, Table, Tag, Typography, message } from 'antd';
import { AuditOutlined, BulbOutlined, CheckCircleOutlined, ClockCircleOutlined, EditOutlined, FileTextOutlined } from '@ant-design/icons';
import { api, Guide, Review, ReviewDomain, Template } from '../api/client';
import { useAuthStore } from '../store/auth';

const { Title, Paragraph } = Typography;

interface ReviewTask {
  guide: Guide;
  review: Review;
}

const statusText: Record<string, string> = {
  draft: '草稿',
  reviewing: '评审中',
  approved: '已通过',
  archived: '已归档',
  pending: '待处理',
  rejected: '已驳回',
};

export default function Dashboard() {
  const user = useAuthStore((s) => s.user);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [guides, setGuides] = useState<Guide[]>([]);
  const [reviewsByGuide, setReviewsByGuide] = useState<Record<string, Review[]>>({});
  const [domains, setDomains] = useState<ReviewDomain[]>([]);
  const [myDomainIds, setMyDomainIds] = useState<string[]>([]);
  const [issues, setIssues] = useState(0);
  const [rules, setRules] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const [tpls, gs, ds, myDs, metrics] = await Promise.all([
          api.listTemplates(),
          api.listGuides(),
          api.listReviewDomains(),
          api.getMyDomains(),
          api.metrics(),
        ]);
        const pairs = await Promise.all(gs.map(async (g) => [g.id, await api.listReviews(g.id)] as const));
        setTemplates(tpls);
        setGuides(gs);
        setDomains(ds);
        setMyDomainIds(myDs.domainIds || []);
        setReviewsByGuide(Object.fromEntries(pairs));
        setIssues(metrics.issueCount);
        setRules(metrics.ruleCount);
      } catch (e: any) {
        message.error(e.message);
      } finally {
        setLoading(false);
      }
    };
    load();
  }, []);

  const myDomains = useMemo(
    () => myDomainIds.map((id) => domains.find((d) => d.id === id)?.name || id),
    [domains, myDomainIds]
  );

  const myReviewTasks = useMemo<ReviewTask[]>(() => {
    if (!user) return [];
    return guides.flatMap((guide) => (reviewsByGuide[guide.id] || [])
      .filter((review) => review.reviewer === user.username && review.status === 'pending')
      .map((review) => ({ guide, review })));
  }, [guides, reviewsByGuide, user]);

  const mySubmittedPending = useMemo(() => {
    if (!user) return [];
    return guides.filter((guide) => {
      if (guide.createdAt && guide.createdBy !== undefined && guide.createdBy !== user.username) return false;
      if (guide.createdBy !== user.username) return false;
      const reviews = reviewsByGuide[guide.id] || [];
      return guide.status === 'reviewing' || reviews.some((review) => review.status === 'pending');
    });
  }, [guides, reviewsByGuide, user]);

  const cardStyle = { background: 'var(--bg-container)', borderColor: 'var(--border-color)' };

  return (
    <div>
      <Title level={3} style={{ color: 'var(--text-primary)', marginBottom: 4 }}>概览</Title>
      <Paragraph style={{ color: 'var(--text-secondary)' }}>
        这里汇总平台模板、我的领域和当前需要处理的评审事项。
      </Paragraph>

      <Row gutter={16}>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="模板总数" value={templates.length} prefix={<FileTextOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="待我评审" value={myReviewTasks.length} prefix={<ClockCircleOutlined />} valueStyle={{ color: 'var(--color-warning)' }} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="我提交待评审" value={mySubmittedPending.length} prefix={<EditOutlined />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={cardStyle}>
            <Statistic title="规则 / 问题" value={`${rules} / ${issues}`} prefix={<BulbOutlined />} valueStyle={{ color: 'var(--color-primary)' }} />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={14}>
          <Card title="待我评审" loading={loading} style={cardStyle}>
            <Table
              rowKey={(r) => r.review.id}
              dataSource={myReviewTasks}
              pagination={false}
              locale={{ emptyText: <Empty description="暂无待评审事项" /> }}
              columns={[
                { title: '评审材料', render: (_, r) => r.guide.title },
                { title: '风险', width: 90, render: (_, r) => <Tag>{r.guide.riskLevel}</Tag> },
                { title: '状态', width: 100, render: (_, r) => <Tag color="processing">{statusText[r.review.status] || r.review.status}</Tag> },
                { title: '创建时间', width: 180, render: (_, r) => r.review.createdAt?.slice(0, 19).replace('T', ' ') },
              ]}
            />
          </Card>
        </Col>
        <Col span={10}>
          <Card title="我的领域" loading={loading} style={cardStyle}>
            {myDomains.length > 0 ? (
              <Space wrap>{myDomains.map((name) => <Tag key={name}>{name}</Tag>)}</Space>
            ) : (
              <Empty description="尚未分配领域" />
            )}
          </Card>
          <Card title="我提交后待评审" loading={loading} style={{ ...cardStyle, marginTop: 16 }}>
            <List
              dataSource={mySubmittedPending}
              locale={{ emptyText: <Empty description="暂无提交后的待评审材料" /> }}
              renderItem={(item) => (
                <List.Item>
                  <List.Item.Meta
                    avatar={<AuditOutlined style={{ color: 'var(--color-primary)' }} />}
                    title={item.title}
                    description={<Space wrap><Tag>{statusText[item.status] || item.status}</Tag><span>{item.updatedAt?.slice(0, 19).replace('T', ' ')}</span></Space>}
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>
      </Row>

      <Card title="模板清单" loading={loading} style={{ ...cardStyle, marginTop: 16 }}>
        <Table
          rowKey="id"
          dataSource={templates}
          pagination={{ pageSize: 6 }}
          locale={{ emptyText: <Empty description="暂无模板" /> }}
          columns={[
            { title: '模板', dataIndex: 'name' },
            { title: '分类', dataIndex: 'category', width: 140, render: (v) => v || '未分类' },
            { title: '质量分', dataIndex: 'qualityScore', width: 110 },
            { title: '使用次数', dataIndex: 'usageCount', width: 110 },
            { title: '状态', dataIndex: 'status', width: 100, render: (v) => <Tag icon={<CheckCircleOutlined />}>{v}</Tag> },
          ]}
        />
      </Card>
    </div>
  );
}
