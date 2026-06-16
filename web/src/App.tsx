import { useEffect } from 'react';
import { Navigate, Routes, Route } from 'react-router-dom';
import { Spin } from 'antd';
import MainLayout from './layouts/MainLayout';
import AuthPage from './pages/Auth';
import Dashboard from './pages/Dashboard';
import Templates from './pages/Templates';
import Guides from './pages/Guides';
import GuideGenerate from './pages/GuideGenerate';
import Reviews from './pages/Reviews';
import Knowledge from './pages/Knowledge';
import Users from './pages/Users';
import Settings from './pages/Settings';
import { useAuthStore } from './store/auth';

export default function App() {
  const { token, user, restoring, restore } = useAuthStore();

  useEffect(() => {
    restore();
  }, [restore]);

  if (restoring) {
    return (
      <div style={{ height: '100vh', display: 'grid', placeItems: 'center', background: 'var(--bg-base)' }}>
        <Spin />
      </div>
    );
  }

  if (!token || !user) return <AuthPage />;

  return (
    <Routes>
      <Route path="/" element={<MainLayout />}>
        <Route index element={<Dashboard />} />
        <Route path="templates" element={<Templates />} />
        <Route path="guides" element={<Guides />} />
        <Route path="guides/new" element={user.role === 'readonly' ? <Navigate to="/guides" replace /> : <GuideGenerate />} />
        <Route path="reviews" element={<Reviews />} />
        <Route path="knowledge" element={<Knowledge />} />
        <Route path="users" element={user.role === 'admin' ? <Users /> : <Navigate to="/" replace />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  );
}
