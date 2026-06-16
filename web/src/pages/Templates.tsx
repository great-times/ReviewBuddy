import { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, List, Modal, Space, Table, Tag, Typography, message } from 'antd';
import { FolderAddOutlined, PlusOutlined } from '@ant-design/icons';
import Editor from '@monaco-editor/react';
import { api, Template, TemplateLibrary } from '../api/client';
import { useAuthStore } from '../store/auth';

const { Title, Paragraph, Text } = Typography;

export default function Templates() {
  const [libraries, setLibraries] = useState<TemplateLibrary[]>([]);
  const [selectedLibraryId, setSelectedLibraryId] = useState('');
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(false);
  const [libraryOpen, setLibraryOpen] = useState(false);
  const [templateOpen, setTemplateOpen] = useState(false);
  const [editing, setEditing] = useState<Partial<Template> | null>(null);
  const [libraryForm] = Form.useForm<Partial<TemplateLibrary>>();
  const canWrite = useAuthStore((s) => s.user?.role !== 'readonly');

  const selectedLibrary = useMemo(
    () => libraries.find((l) => l.id === selectedLibraryId),
    [libraries, selectedLibraryId]
  );

  const loadLibraries = () => {
    api.listTemplateLibraries()
      .then((items) => {
        setLibraries(items);
        setSelectedLibraryId((current) => current || items[0]?.id || '');
      })
      .catch((e) => message.error(e.message));
  };

  const loadTemplates = (libraryId = selectedLibraryId) => {
    if (!libraryId) return;
    setLoading(true);
    api.listTemplates(libraryId).then(setTemplates).catch((e) => message.error(e.message)).finally(() => setLoading(false));
  };

  useEffect(loadLibraries, []);
  useEffect(() => {
    loadTemplates(selectedLibraryId);
  }, [selectedLibraryId]);

  const saveLibrary = async () => {
    const values = await libraryForm.validateFields();
    try {
      const created = await api.createTemplateLibrary(values);
      message.success('模板库已创建');
      setLibraryOpen(false);
      libraryForm.resetFields();
      await api.listTemplateLibraries().then((items) => {
        setLibraries(items);
        setSelectedLibraryId(created.id);
      });
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const openTemplate = (t?: Template) => {
    setEditing(t || { libraryId: selectedLibraryId, category: selectedLibrary?.name || '标准', content: '', variables: [] });
    setTemplateOpen(true);
  };

  const saveTemplate = async () => {
    if (!editing?.name) return message.warning('请填写模板名称');
    try {
      const payload = { ...editing, libraryId: editing.libraryId || selectedLibraryId };
      if (editing.id) await api.updateTemplate(editing.id, payload);
      else await api.createTemplate(payload);
      message.success('模板已保存');
      setTemplateOpen(false);
      loadTemplates();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <div>
          <Title level={3} style={{ color: 'var(--text-primary)', marginBottom: 4 }}>模板库</Title>
          <Paragraph style={{ color: 'var(--text-secondary)', marginBottom: 0 }}>
            先创建模板库，再在对应模板库中维护评审模板。
          </Paragraph>
        </div>
        {canWrite && (
          <Button icon={<FolderAddOutlined />} onClick={() => setLibraryOpen(true)}>
            新建模板库
          </Button>
        )}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '280px 1fr', gap: 16 }}>
        <Card title="模板库" style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}>
          <List
            dataSource={libraries}
            locale={{ emptyText: '暂无模板库' }}
            renderItem={(item) => (
              <List.Item
                style={{
                  cursor: 'pointer',
                  paddingInline: 10,
                  borderRadius: 8,
                  background: item.id === selectedLibraryId ? 'var(--color-primary-opacity-10)' : 'transparent',
                }}
                onClick={() => setSelectedLibraryId(item.id)}
              >
                <Space direction="vertical" size={2}>
                  <Text strong>{item.name}</Text>
                  {item.description && <Text type="secondary" style={{ fontSize: 12 }}>{item.description}</Text>}
                </Space>
              </List.Item>
            )}
          />
        </Card>

        <Card
          title={selectedLibrary?.name || '模板'}
          extra={canWrite && selectedLibraryId ? (
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openTemplate()}>
              新建模板
            </Button>
          ) : null}
          style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}
        >
          <Table
            rowKey="id"
            loading={loading}
            dataSource={templates}
            pagination={false}
            columns={[
              { title: '名称', dataIndex: 'name' },
              { title: '分类', dataIndex: 'category', render: (c) => <Tag color="green">{c}</Tag> },
              { title: '描述', dataIndex: 'description', ellipsis: true },
              { title: '版本', dataIndex: 'currentVersion', width: 80 },
              { title: '使用次数', dataIndex: 'usageCount', width: 90 },
              ...(canWrite ? [{
                title: '操作',
                width: 100,
                render: (_: unknown, r: Template) => (
                  <Button size="small" onClick={() => openTemplate(r)}>编辑</Button>
                ),
              }] : []),
            ]}
          />
        </Card>
      </div>

      <Modal title="新建模板库" open={libraryOpen} onCancel={() => setLibraryOpen(false)} onOk={saveLibrary} okText="创建" cancelText="取消">
        <Form form={libraryForm} layout="vertical">
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入模板库名称' }]}>
            <Input placeholder="如：架构评审模板库" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="说明模板库适用的评审域或团队" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editing?.id ? '编辑模板' : '新建模板'}
        open={templateOpen}
        onCancel={() => setTemplateOpen(false)}
        onOk={saveTemplate}
        width={820}
        okText="保存"
        cancelText="取消"
      >
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Input
            placeholder="模板名称，如：架构方案评审模板"
            value={editing?.name}
            onChange={(e) => setEditing({ ...editing, name: e.target.value })}
          />
          <Input
            placeholder="分类，如：架构 / 设计 / 测试 / 安全"
            value={editing?.category}
            onChange={(e) => setEditing({ ...editing, category: e.target.value })}
          />
          <Input
            placeholder="描述"
            value={editing?.description}
            onChange={(e) => setEditing({ ...editing, description: e.target.value })}
          />
          <div style={{ border: '1px solid var(--border-color)', borderRadius: 8, overflow: 'hidden' }}>
            <Editor
              height="320px"
              defaultLanguage="markdown"
              value={editing?.content || ''}
              onChange={(v) => setEditing({ ...editing, content: v || '' })}
              options={{ minimap: { enabled: false }, fontSize: 13, wordWrap: 'on' }}
            />
          </div>
        </Space>
      </Modal>
    </div>
  );
}
