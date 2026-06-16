import { useState } from 'react';
import { Button, Card, Form, Input, Segmented, Typography, message } from 'antd';
import { LoginOutlined, UserAddOutlined } from '@ant-design/icons';
import Logo from '../components/Logo';
import { useAuthStore } from '../store/auth';

const { Paragraph, Text } = Typography;

export default function AuthPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [loading, setLoading] = useState(false);
  const login = useAuthStore((s) => s.login);
  const register = useAuthStore((s) => s.register);

  const submit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      if (mode === 'login') {
        await login(values.username, values.password);
      } else {
        await register(values.username, values.password);
      }
    } catch (e: any) {
      message.error(e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page">
      <Card className="auth-panel">
        <div style={{ marginBottom: 24 }}>
          <Logo size={34} />
          <Paragraph style={{ color: 'var(--text-secondary)', margin: '10px 0 0' }}>
            登录后进入智能评审助手。注册不选择角色，默认只读，由管理员在用户管理中分配。
          </Paragraph>
        </div>

        <Segmented
          block
          value={mode}
          onChange={(v) => setMode(v as 'login' | 'register')}
          options={[
            { label: '登录', value: 'login', icon: <LoginOutlined /> },
            { label: '注册', value: 'register', icon: <UserAddOutlined /> },
          ]}
          style={{ marginBottom: 20 }}
        />

        <Form layout="vertical" onFinish={submit}>
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input autoComplete="username" placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password autoComplete={mode === 'login' ? 'current-password' : 'new-password'} placeholder="至少 4 位" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={loading} icon={mode === 'login' ? <LoginOutlined /> : <UserAddOutlined />}>
            {mode === 'login' ? '登录' : '注册'}
          </Button>
        </Form>

        {mode === 'register' && (
          <Text type="secondary" style={{ display: 'block', marginTop: 14, fontSize: 12 }}>
            首个注册账号会自动成为管理员；之后注册的账号默认只读，可查看内容但不能增删改。
          </Text>
        )}
      </Card>
    </div>
  );
}
