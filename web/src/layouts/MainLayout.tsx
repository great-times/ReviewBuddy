import { Layout, Menu, Dropdown, Tooltip, Tag, Typography } from 'antd';
import {
  DashboardOutlined,
  FileTextOutlined,
  EditOutlined,
  AuditOutlined,
  BulbOutlined,
  BgColorsOutlined,
  SettingOutlined,
  TeamOutlined,
  LogoutOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useLocation, useNavigate, Outlet } from 'react-router-dom';
import Logo from '../components/Logo';
import { useAuthStore } from '../store/auth';
import { useThemeStore } from '../store/theme';
import { themeList } from '../themes';
import { hasRole, roleColor, userRoles } from '../utils/roles';

const { Sider, Header, Content } = Layout;

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '概览', title: '评审助手总览与度量' },
  { key: '/templates', icon: <FileTextOutlined />, label: '模板库', title: '按库管理评审模板，沉淀最佳实践' },
  { key: '/guides', icon: <EditOutlined />, label: '评审材料', title: 'AI 辅助生成与管理评审材料' },
  { key: '/reviews', icon: <AuditOutlined />, label: '评审工作台', title: 'Hermes AI 预审 + 人工评审意见' },
  { key: '/knowledge', icon: <BulbOutlined />, label: '规则沉淀', title: '问题沉淀、规则提炼、模板反哺' },
  { key: '/users', icon: <TeamOutlined />, label: '用户管理', title: '配置管理员、开发、运维、测试、架构、设计角色' },
  { key: '/settings', icon: <SettingOutlined />, label: '设置', title: 'Agent 与模型配置' },
];

export default function MainLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { currentTheme, setTheme } = useThemeStore();
  const { user, logout } = useAuthStore();
  const visibleMenuItems = hasRole(user, 'admin') ? menuItems : menuItems.filter((m) => m.key !== '/users');
  const roles = userRoles(user);
  const roleLabel = roles.length > 0 ? roles.join(' / ') : '';

  const selectedKey =
    visibleMenuItems
      .map((m) => m.key)
      .filter((k) => k !== '/' && location.pathname.startsWith(k))
      .sort((a, b) => b.length - a.length)[0] || '/';

  return (
    <Layout style={{ height: '100vh' }}>
      <Sider theme="light" style={{ background: 'var(--bg-sidebar)', borderRight: '1px solid var(--border-color)' }}>
        <div style={{ padding: '18px 16px' }}>
          <Logo size={30} />
          <div style={{ marginTop: 8, fontSize: 12, color: 'var(--text-secondary)', lineHeight: 1.5 }}>
            智能评审助手
          </div>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={visibleMenuItems}
          style={{ background: 'transparent', borderInlineEnd: 'none' }}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            background: 'var(--bg-container)',
            borderBottom: '1px solid var(--border-color)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'flex-end',
            gap: 16,
            padding: '0 20px',
          }}
        >
          {user && (
            <Dropdown
              menu={{
                items: [
                  { key: 'name', label: <Typography.Text type="secondary">{user.username}</Typography.Text>, disabled: true, icon: <UserOutlined /> },
                  { key: 'role', label: <span>{roles.map((role) => <Tag key={role} color={roleColor(role)}>{role}</Tag>)}</span>, disabled: true },
                  { type: 'divider' },
                  { key: 'logout', label: '退出登录', icon: <LogoutOutlined /> },
                ],
                onClick: ({ key }) => {
                  if (key === 'logout') logout();
                },
              }}
            >
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8, cursor: 'pointer', color: 'var(--text-primary)' }}>
                <UserOutlined />
                <span>{user.username}</span>
                <Tag color={hasRole(user, 'admin') ? 'red' : roles.every((role) => role === 'readonly') ? 'default' : 'blue'} style={{ marginInlineEnd: 0 }}>{roleLabel}</Tag>
              </span>
            </Dropdown>
          )}
          <Dropdown
            menu={{
              items: themeList.map((t) => ({
                key: t.name,
                label: (
                  <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span
                      style={{
                        width: 12,
                        height: 12,
                        borderRadius: '50%',
                        background: t.color,
                        display: 'inline-block',
                      }}
                    />
                    {t.label}
                  </span>
                ),
              })),
              selectedKeys: [currentTheme],
              onClick: ({ key }) => setTheme(key as never),
            }}
          >
            <Tooltip title="切换主题">
              <BgColorsOutlined style={{ fontSize: 18, cursor: 'pointer', color: 'var(--text-primary)' }} />
            </Tooltip>
          </Dropdown>
        </Header>
        <Content style={{ padding: 20, overflow: 'auto', background: 'var(--bg-base)' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
