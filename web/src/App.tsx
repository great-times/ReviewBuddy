import { lazy, Suspense, useEffect } from 'react';
import { Navigate, Routes, Route } from 'react-router-dom';
import { Spin } from 'antd';
import MainLayout from './layouts/MainLayout';
import AuthPage from './pages/Auth';
import { useAuthStore } from './store/auth';
import { canWrite, isActiveRole } from './utils/roles';

const Dashboard = lazy(() => import('./pages/Dashboard'));
const Templates = lazy(() => import('./pages/Templates'));
const Guides = lazy(() => import('./pages/Guides'));
const GuideGenerate = lazy(() => import('./pages/GuideGenerate'));
const Reviews = lazy(() => import('./pages/Reviews'));
const Knowledge = lazy(() => import('./pages/Knowledge'));
const Users = lazy(() => import('./pages/Users'));
const Settings = lazy(() => import('./pages/Settings'));

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
    <Suspense fallback={<div style={{ height: '100%', display: 'grid', placeItems: 'center' }}><Spin /></div>}>
      <Routes>
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="templates" element={<Templates />} />
          <Route path="guides" element={<Guides />} />
          <Route path="guides/new" element={canWrite(user) ? <GuideGenerate /> : <Navigate to="/guides" replace />} />
          <Route path="reviews" element={<Reviews />} />
          <Route path="knowledge" element={<Knowledge />} />
          <Route path="users" element={isActiveRole(user, 'admin') ? <Users /> : <Navigate to="/" replace />} />
          <Route path="settings" element={<Settings />} />
        </Route>
      </Routes>
    </Suspense>
  );
}
