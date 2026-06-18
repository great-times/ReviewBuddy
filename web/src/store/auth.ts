import { create } from 'zustand';
import { ACTIVE_ROLE_KEY, api, EXPIRES_KEY, LoginResult, TOKEN_KEY, User } from '../api/client';
import { effectiveRole, userRoles } from '../utils/roles';

function storedToken() {
  const token = localStorage.getItem(TOKEN_KEY);
  const expires = localStorage.getItem(EXPIRES_KEY);
  if (!token || !expires || Date.now() > new Date(expires).getTime()) {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(EXPIRES_KEY);
    return null;
  }
  return token;
}

interface AuthState {
  token: string | null;
  user: User | null;
  activeRole: string;
  restoring: boolean;
  applyLogin: (result: LoginResult) => void;
  setActiveRole: (role: string) => void;
  restore: () => Promise<void>;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string) => Promise<LoginResult>;
  logout: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: storedToken(),
  user: null,
  activeRole: localStorage.getItem(ACTIVE_ROLE_KEY) || '',
  restoring: false,

  applyLogin: (result) => {
    if (!result.token || !result.expiresAt) return;
    localStorage.setItem(TOKEN_KEY, result.token);
    localStorage.setItem(EXPIRES_KEY, result.expiresAt);
    const roles = userRoles(result.user);
    const nextRole = roles.includes(localStorage.getItem(ACTIVE_ROLE_KEY) || '')
      ? localStorage.getItem(ACTIVE_ROLE_KEY) || ''
      : effectiveRole(result.user);
    if (nextRole) localStorage.setItem(ACTIVE_ROLE_KEY, nextRole);
    set({ token: result.token, user: { ...result.user, role: nextRole || result.user.role }, activeRole: nextRole });
  },

  setActiveRole: (role) => {
    const user = get().user;
    if (!user || !userRoles(user).includes(role)) return;
    localStorage.setItem(ACTIVE_ROLE_KEY, role);
    set({ activeRole: role, user: { ...user, role } });
  },

  restore: async () => {
    const token = storedToken();
    if (!token) {
      set({ token: null, user: null, restoring: false });
      return;
    }
    set({ token, restoring: true });
    try {
      const user = await api.me();
      const nextRole = effectiveRole(user);
      if (nextRole) localStorage.setItem(ACTIVE_ROLE_KEY, nextRole);
      set({ token, user, activeRole: nextRole });
    } catch {
      localStorage.removeItem(ACTIVE_ROLE_KEY);
      try {
        const user = await api.me();
        const nextRole = effectiveRole(user);
        if (nextRole) localStorage.setItem(ACTIVE_ROLE_KEY, nextRole);
        set({ token, user, activeRole: nextRole });
      } catch {
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(EXPIRES_KEY);
        set({ token: null, user: null, activeRole: '' });
      }
    } finally {
      set({ restoring: false });
    }
  },

  login: async (username, password) => {
    get().applyLogin(await api.login(username, password));
  },

  register: async (username, password) => {
    const result = await api.register(username, password);
    get().applyLogin(result);
    return result;
  },

  logout: async () => {
    try {
      await api.logout();
    } finally {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(EXPIRES_KEY);
      localStorage.removeItem(ACTIVE_ROLE_KEY);
      set({ token: null, user: null, activeRole: '' });
    }
  },
}));
