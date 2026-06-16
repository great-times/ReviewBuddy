import { useEffect, useState } from 'react';
import { Button, Card, Table, Tag, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { api, Guide } from '../api/client';
import { useAuthStore } from '../store/auth';

const { Title } = Typography;

const statusColor: Record<string, string> = {
  draft: 'default',
  reviewing: 'processing',
  approved: 'success',
  archived: 'purple',
};
const statusLabel: Record<string, string> = {
  draft: '草稿',
  reviewing: '评审中',
  approved: '已通过',
  archived: '已归档',
};

export default function Guides() {
  const [list, setList] = useState<Guide[]>([]);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const canWrite = useAuthStore((s) => s.user?.role !== 'readonly');

  const load = () => {
    setLoading(true);
    api.listGuides().then(setList).catch((e) => message.error(e.message)).finally(() => setLoading(false));
  };
  useEffect(load, []);

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Title level={3} style={{ color: 'var(--text-primary)' }}>评审材料</Title>
        {canWrite && (
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/guides/new')}>
          AI 生成材料
          </Button>
        )}
      </div>
      <Card style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={list}
          pagination={false}
          columns={[
            { title: '标题', dataIndex: 'title' },
            {
              title: '状态',
              dataIndex: 'status',
              width: 110,
              render: (s) => <Tag color={statusColor[s]}>{statusLabel[s] || s}</Tag>,
            },
            { title: '风险', dataIndex: 'riskLevel', width: 90 },
            { title: '版本', dataIndex: 'currentVersion', width: 80 },
            { title: '更新时间', dataIndex: 'updatedAt', width: 200 },
          ]}
        />
      </Card>
    </div>
  );
}
