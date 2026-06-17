import { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Modal, Popconfirm, Select, Space, Table, Tabs, Tag, Typography, message } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { api, DomainRoleUsers, ReviewDomain, ReviewRole, ReviewScenario, User, UserRole } from '../api/client';

const { Title, Paragraph } = Typography;

const systemRoleNames: Record<string, string> = { admin: '管理员', readonly: '只读' };
const systemRoleColors: Record<string, string> = { admin: 'red', readonly: 'default' };

export default function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<ReviewRole[]>([]);
  const [domains, setDomains] = useState<ReviewDomain[]>([]);
  const [scenarios, setScenarios] = useState<ReviewScenario[]>([]);
  const [roleUsers, setRoleUsers] = useState<DomainRoleUsers[]>([]);
  const [userDomains, setUserDomains] = useState<Record<string, string[]>>({});
  const [activeDomainId, setActiveDomainId] = useState('');
  const [loading, setLoading] = useState(false);
  const [userOpen, setUserOpen] = useState(false);
  const [roleOpen, setRoleOpen] = useState(false);
  const [domainOpen, setDomainOpen] = useState(false);
  const [scenarioOpen, setScenarioOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<Partial<User> | null>(null);
  const [editingRole, setEditingRole] = useState<Partial<ReviewRole> | null>(null);
  const [editingDomain, setEditingDomain] = useState<Partial<ReviewDomain> | null>(null);
  const [editingScenario, setEditingScenario] = useState<Partial<ReviewScenario> | null>(null);
  const [userForm] = Form.useForm<Partial<User> & { domainIds?: string[] }>();
  const [roleForm] = Form.useForm<Partial<ReviewRole>>();
  const [domainForm] = Form.useForm<Partial<ReviewDomain>>();
  const [scenarioForm] = Form.useForm<Partial<ReviewScenario>>();

  const roleMeta = useMemo(() => Object.fromEntries(roles.map((r) => [r.key, r])), [roles]);
  const roleOptions = roles.map((r) => ({ value: r.key, label: r.name }));
  const assignableRoles = roles.filter((r) => r.key !== 'admin' && r.key !== 'readonly');
  const reviewers = users.filter((u) => u.role !== 'readonly');

  const load = async () => {
    setLoading(true);
    try {
      const [u, r, d, s] = await Promise.all([
        api.listUsers(),
        api.listReviewRoles(),
        api.listReviewDomains(),
        api.listReviewScenarios(),
      ]);
      setUsers(u);
      setRoles(r);
      setDomains(d);
      setScenarios(s);
      const pairs = await Promise.all(u.map(async (item) => [item.id, (await api.getUserDomains(item.id)).domainIds] as const));
      setUserDomains(Object.fromEntries(pairs));
      const nextDomain = activeDomainId || d[0]?.id || '';
      setActiveDomainId(nextDomain);
      if (nextDomain) setRoleUsers(await api.listDomainRoleUsers(nextDomain));
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const loadRoleUsers = async (domainId: string) => {
    setActiveDomainId(domainId);
    if (!domainId) return setRoleUsers([]);
    try {
      setRoleUsers(await api.listDomainRoleUsers(domainId));
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const roleName = (key: string) => roleMeta[key]?.name || systemRoleNames[key] || key;
  const roleTag = (key: string) => <Tag color={systemRoleColors[key] || 'blue'}>{roleName(key)}</Tag>;

  const editUser = (u?: User) => {
    const value = u ? { ...u, domainIds: userDomains[u.id] || [] } : { role: 'readonly' as UserRole, domainIds: [] };
    setEditingUser(value);
    userForm.setFieldsValue(value);
    setUserOpen(true);
  };

  const saveUser = async () => {
    const values = await userForm.validateFields();
    try {
      const { domainIds = [], ...userValues } = values;
      const saved = editingUser?.id
        ? await api.updateUser(editingUser.id, { ...editingUser, ...userValues })
        : await api.createUser(userValues);
      await api.updateUserDomains(saved.id, domainIds);
      message.success('用户已保存');
      setUserOpen(false);
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const editRole = (role?: ReviewRole) => {
    setEditingRole(role || {});
    roleForm.setFieldsValue(role || {});
    setRoleOpen(true);
  };

  const saveRole = async () => {
    const values = await roleForm.validateFields();
    try {
      if (editingRole?.key) await api.updateReviewRole(editingRole.key, values);
      else await api.createReviewRole(values);
      message.success('角色已保存');
      setRoleOpen(false);
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const editDomain = (domain?: ReviewDomain) => {
    setEditingDomain(domain || {});
    domainForm.setFieldsValue(domain || {});
    setDomainOpen(true);
  };

  const saveDomain = async () => {
    const values = await domainForm.validateFields();
    try {
      if (editingDomain?.id) await api.updateReviewDomain(editingDomain.id, values);
      else await api.createReviewDomain(values);
      message.success('领域已保存');
      setDomainOpen(false);
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const editScenario = (scenario?: ReviewScenario) => {
    const value = scenario || { roleKeys: [] };
    setEditingScenario(value);
    scenarioForm.setFieldsValue(value);
    setScenarioOpen(true);
  };

  const saveScenario = async () => {
    const values = await scenarioForm.validateFields();
    try {
      if (editingScenario?.id) await api.updateReviewScenario(editingScenario.id, values);
      else await api.createReviewScenario(values);
      message.success('场景已保存');
      setScenarioOpen(false);
      load();
    } catch (e: any) {
      message.error(e.message);
    }
  };

  const saveDomainUsers = async (roleKey: string, userIds: string[]) => {
    if (!activeDomainId) return;
    try {
      await api.updateDomainRoleUsers(activeDomainId, roleKey, userIds);
      setRoleUsers(await api.listDomainRoleUsers(activeDomainId));
      message.success('默认人员已保存');
    } catch (e: any) {
      message.error(e.message);
    }
  };

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <Title level={3} style={{ color: 'var(--text-primary)', marginBottom: 4 }}>用户与评审配置</Title>
          <Paragraph style={{ color: 'var(--text-secondary)', marginBottom: 0 }}>
            角色可自定义；新注册用户默认只读，由管理员分配角色。领域中维护每个角色的默认人员，场景中配置需要哪些角色审批。
          </Paragraph>
        </div>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => editUser()}>新增用户</Button>
      </div>

      <Tabs
        style={{ marginTop: 12 }}
        items={[
          {
            key: 'users',
            label: '用户与角色',
            children: (
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                <Card style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}>
                  <Table
                    rowKey="id"
                    loading={loading}
                    dataSource={users}
                    pagination={false}
                    columns={[
                      { title: '姓名', dataIndex: 'username' },
                      { title: '角色', dataIndex: 'role', width: 140, render: roleTag },
                      {
                        title: '所属领域',
                        dataIndex: 'id',
                        render: (id: string) => {
                          const names = (userDomains[id] || []).map((domainId) => domains.find((d) => d.id === domainId)?.name || domainId);
                          return names.length > 0 ? <Space wrap>{names.map((name) => <Tag key={name}>{name}</Tag>)}</Space> : <Typography.Text type="secondary">未分配</Typography.Text>;
                        },
                      },
                      {
                        title: '操作',
                        width: 150,
                        render: (_, r) => (
                          <Space>
                            <Button size="small" onClick={() => editUser(r)}>编辑</Button>
                            <Popconfirm title="确认删除该用户？" okText="删除" cancelText="取消" onConfirm={async () => {
                              await api.deleteUser(r.id);
                              message.success('用户已删除');
                              load();
                            }}>
                              <Button size="small" danger icon={<DeleteOutlined />} />
                            </Popconfirm>
                          </Space>
                        ),
                      },
                    ]}
                  />
                </Card>

                <Card
                  title="角色类型"
                  extra={<Button icon={<PlusOutlined />} onClick={() => editRole()}>新增角色</Button>}
                  style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}
                >
                  <Table
                    rowKey="key"
                    dataSource={roles}
                    pagination={false}
                    columns={[
                      { title: '角色', dataIndex: 'name', width: 160, render: (_, r) => roleTag(r.key) },
                      { title: '编码', dataIndex: 'key', width: 160 },
                      { title: '说明', dataIndex: 'description' },
                      {
                        title: '操作',
                        width: 150,
                        render: (_, r) => r.system ? <Tag>系统角色</Tag> : (
                          <Space>
                            <Button size="small" onClick={() => editRole(r)}>编辑</Button>
                            <Popconfirm title="确认删除该角色？" okText="删除" cancelText="取消" onConfirm={async () => {
                              await api.deleteReviewRole(r.key);
                              message.success('角色已删除');
                              load();
                            }}>
                              <Button size="small" danger icon={<DeleteOutlined />} />
                            </Popconfirm>
                          </Space>
                        ),
                      },
                    ]}
                  />
                </Card>
              </Space>
            ),
          },
          {
            key: 'domains',
            label: '领域默认人员',
            children: (
              <Card
                title="领域角色人员"
                extra={<Button icon={<PlusOutlined />} onClick={() => editDomain()}>新增领域</Button>}
                style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}
              >
                <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                  <Space wrap>
                    <Select
                      style={{ minWidth: 220 }}
                      value={activeDomainId || undefined}
                      placeholder="选择领域"
                      options={domains.map((d) => ({ value: d.id, label: d.name }))}
                      onChange={loadRoleUsers}
                    />
                    {domains.find((d) => d.id === activeDomainId) && (
                      <Button onClick={() => editDomain(domains.find((d) => d.id === activeDomainId))}>编辑领域</Button>
                    )}
                  </Space>
                  <Table
                    rowKey="key"
                    dataSource={assignableRoles}
                    pagination={false}
                    columns={[
                      { title: '角色', dataIndex: 'key', width: 160, render: roleTag },
                      { title: '说明', dataIndex: 'description', width: 240 },
                      {
                        title: '默认人员',
                        render: (_, r) => {
                          const saved = roleUsers.find((x) => x.roleKey === r.key)?.userIds || [];
                          return (
                            <Select
                              mode="multiple"
                              style={{ width: '100%' }}
                              value={saved}
                              placeholder={`选择${r.name}默认人员`}
                              options={reviewers.filter((u) => u.role === r.key).map((u) => ({ value: u.id, label: u.username }))}
                              onChange={(ids) => saveDomainUsers(r.key, ids)}
                            />
                          );
                        },
                      },
                    ]}
                  />
                </Space>
              </Card>
            ),
          },
          {
            key: 'scenarios',
            label: '审批场景',
            children: (
              <Card
                title="场景角色要求"
                extra={<Button icon={<PlusOutlined />} onClick={() => editScenario()}>新增场景</Button>}
                style={{ background: 'var(--bg-container)', borderColor: 'var(--border-color)' }}
              >
                <Table
                  rowKey="id"
                  dataSource={scenarios}
                  pagination={false}
                  columns={[
                    { title: '场景', dataIndex: 'name', width: 180 },
                    { title: '说明', dataIndex: 'description' },
                    { title: '需要审批角色', dataIndex: 'roleKeys', render: (keys: string[]) => <Space wrap>{keys.map((k) => roleTag(k))}</Space> },
                    {
                      title: '操作',
                      width: 150,
                      render: (_, r) => (
                        <Space>
                          <Button size="small" onClick={() => editScenario(r)}>编辑</Button>
                          <Popconfirm title="确认删除该场景？" okText="删除" cancelText="取消" onConfirm={async () => {
                            await api.deleteReviewScenario(r.id);
                            message.success('场景已删除');
                            load();
                          }}>
                            <Button size="small" danger disabled={r.id === 'standard'} icon={<DeleteOutlined />} />
                          </Popconfirm>
                        </Space>
                      ),
                    },
                  ]}
                />
              </Card>
            ),
          },
        ]}
      />

      <Modal title={editingUser?.id ? '编辑用户' : '新增用户'} open={userOpen} onCancel={() => setUserOpen(false)} onOk={saveUser} okText="保存" cancelText="取消">
        <Form form={userForm} layout="vertical">
          <Form.Item name="username" label="姓名" rules={[{ required: true, message: '请输入姓名' }]}>
            <Input placeholder="如：张三" />
          </Form.Item>
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select options={roleOptions} />
          </Form.Item>
          <Form.Item name="domainIds" label="所属领域">
            <Select mode="multiple" options={domains.map((d) => ({ value: d.id, label: d.name }))} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title={editingRole?.key ? '编辑角色' : '新增角色'} open={roleOpen} onCancel={() => setRoleOpen(false)} onOk={saveRole} okText="保存" cancelText="取消">
        <Form form={roleForm} layout="vertical">
          <Form.Item name="name" label="角色名称" rules={[{ required: true, message: '请输入角色名称' }]}>
            <Input placeholder="如：安全、法务、财务" />
          </Form.Item>
          {!editingRole?.key && (
            <Form.Item name="key" label="角色编码" rules={[{ required: true, message: '请输入角色编码' }]}>
              <Input placeholder="如：security，仅支持小写字母、数字和下划线" />
            </Form.Item>
          )}
          <Form.Item name="description" label="说明">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title={editingDomain?.id ? '编辑领域' : '新增领域'} open={domainOpen} onCancel={() => setDomainOpen(false)} onOk={saveDomain} okText="保存" cancelText="取消">
        <Form form={domainForm} layout="vertical">
          <Form.Item name="name" label="领域名称" rules={[{ required: true, message: '请输入领域名称' }]}>
            <Input placeholder="如：交易域、用户增长、基础设施" />
          </Form.Item>
          <Form.Item name="description" label="说明">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal title={editingScenario?.id ? '编辑场景' : '新增场景'} open={scenarioOpen} onCancel={() => setScenarioOpen(false)} onOk={saveScenario} okText="保存" cancelText="取消">
        <Form form={scenarioForm} layout="vertical">
          <Form.Item name="name" label="场景名称" rules={[{ required: true, message: '请输入场景名称' }]}>
            <Input placeholder="如：标准评审、上线评审、安全评审" />
          </Form.Item>
          <Form.Item name="roleKeys" label="需要审批角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select mode="multiple" options={assignableRoles.map((r) => ({ value: r.key, label: r.name }))} />
          </Form.Item>
          <Form.Item name="description" label="说明">
            <Input.TextArea rows={3} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
