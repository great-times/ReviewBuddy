import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tag, Typography, message } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { api, User, UserRole } from '../api/client';

const { Title, Paragraph } = Typography;

const roleOptions: { label: string; value: UserRole; color: string }[] = [
  { label: '管理员', value: 'admin', color: 'red' },
  { label: '只读', value: 'readonly', color: 'default' },
  { label: '开发', value: 'developer', color: 'blue' },
  { label: '运维', value: 'ops', color: 'green' },
  { label: '测试', value: 'tester', color: 'gold' },
  { label: '架构', value: 'architect', color: 'purple' },
  { label: '设计', value: 'designer', color: 'magenta' },
];

const roleMeta = Object.fromEntries(roleOptions.map((r) => [r.value, r]));

export default function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Partial<User> | null>(null);
  const [form] = Form.useForm<Partial<User>>();

  const load = () => {
    setLoading(true);
    api.listUsers().then(setUsers).catch((e) => message.error(e.message)).finally(() => setLoading(false));
  };

  useEffect(load, []);

  const edit = (u?: User) => {
    const value = u || { role: 'readonly' as UserRole };
    setEditing(value);
    form.setFieldsValue(value);
    setOpen(true);
  };

  const save = async () => {
    const values = await form.validateFields();
    try {
      if (editing?.id) await api.updateUser(editing.id, { ...editing, ...values });
      else await api.createUser(values);
      message.success('用户已保存');
      setOpen(false);
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const remove = async (id: string) => {
    try {
      await api.deleteUser(id);
      message.success('用户已删除');
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <Title level={3} style={{ color: 'var(--text-primary)', marginBottom: 4 }}>用户管理</Title>
          <Paragraph style={{ color: 'var(--text-secondary)', marginBottom: 0 }}>
            维护评审参与人及角色；新注册用户默认只读，由管理员分配为对应角色。
          </Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => edit()}>新增用户</Button>
      </div>

      <Card style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)', marginTop: 16 }}>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={users}
          pagination={false}
          columns={[
            { title: '姓名', dataIndex: 'username' },
            {
              title: '角色',
              dataIndex: 'role',
              width: 120,
              render: (r: UserRole) => <Tag color={roleMeta[r]?.color}>{roleMeta[r]?.label || r}</Tag>,
            },
            {
              title: '操作',
              width: 150,
              render: (_, r) => (
                <Space>
                  <Button size="small" onClick={() => edit(r)}>编辑</Button>
                  <Popconfirm title="确认删除该用户？" okText="删除" cancelText="取消" onConfirm={() => remove(r.id)}>
                    <Button size="small" danger icon={<DeleteOutlined />} />
                  </Popconfirm>
                </Space>
              ),
            },
          ]}
        />
      </Card>

      <Modal title={editing?.id ? '编辑用户' : '新增用户'} open={open} onCancel={() => setOpen(false)} onOk={save} okText="保存" cancelText="取消">
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="如：张三" />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select options={roleOptions.map(({ label, value }) => ({ label, value }))} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
