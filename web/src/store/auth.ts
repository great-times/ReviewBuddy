import { create } from 'zustand';
import { api, EXPIRES_KEY, LoginResult, TOKEN_KEY, User } from '../api/client';

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
  restoring: boolean;
  applyLogin: (result: LoginResult) => void;
  restore: () => Promise<void>;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, password: string) => Promise<LoginResult>;
  logout: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: storedToken(),
  user: null,
  restoring: false,

  applyLogin: (result) => {
    if (!result.token || !result.expiresAt) return;
    localStorage.setItem(TOKEN_KEY, result.token);
    localStorage.setItem(EXPIRES_KEY, result.expiresAt);
    set({ token: result.token, user: result.user });
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
      set({ token, user });
    } catch {
      localStorage.removeItem(TOKEN_KEY);
      localStorage.removeItem(EXPIRES_KEY);
      set({ token: null, user: null });
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
      set({ token: null, user: null });
    }
  },
}));
