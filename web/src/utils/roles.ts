import type { User } from '../api/client';

export const userRoles = (user?: Pick<User, 'role' | 'roles'> | null) => {
  if (!user) return [];
  return user.roles?.length ? user.roles : user.role ? [user.role] : [];
};

export const hasRole = (user: Pick<User, 'role' | 'roles'> | null | undefined, role: string) =>
  userRoles(user).includes(role);

export const canWrite = (user: Pick<User, 'role' | 'roles'> | null | undefined) => {
  if (!user) return false;
  const roles = userRoles(user);
  return roles.some((role) => role !== 'readonly');
};

export const roleColor = (role: string) => {
  if (role === 'admin') return 'red';
  if (role === 'readonly') return 'default';
  return 'blue';
};
