import type { User } from '../api/client';

export const userRoles = (user?: Pick<User, 'role' | 'roles'> | null) => {
  if (!user) return [];
  return user.roles?.length ? user.roles : user.role ? [user.role] : [];
};

export const effectiveRole = (user?: Pick<User, 'role' | 'roles'> | null) => {
  if (!user) return '';
  const roles = userRoles(user);
  if (user.role && roles.includes(user.role)) return user.role;
  if (roles.includes('admin')) return 'admin';
  return roles.find((role) => role !== 'readonly') || roles[0] || '';
};

export const hasRole = (user: Pick<User, 'role' | 'roles'> | null | undefined, role: string) =>
  userRoles(user).includes(role);

export const isActiveRole = (user: Pick<User, 'role' | 'roles'> | null | undefined, role: string) =>
  effectiveRole(user) === role;

export const canWrite = (user: Pick<User, 'role' | 'roles'> | null | undefined) => {
  if (!user) return false;
  return effectiveRole(user) !== 'readonly';
};

export const roleColor = (role: string) => {
  if (role === 'admin') return 'red';
  if (role === 'readonly') return 'default';
  return 'blue';
};
