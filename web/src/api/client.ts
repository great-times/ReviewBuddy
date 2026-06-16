import axios from 'axios';

export const http = axios.create({ baseURL: '/api' });

export const TOKEN_KEY = 'reviewbuddy_auth_token';
export const EXPIRES_KEY = 'reviewbuddy_auth_expires_at';

http.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (res) => res,
  (err) => {
    const msg = err?.response?.data?.error || err.message || '请求失败';
    return Promise.reject(new Error(msg));
  }
);

export interface Template {
  id: string;
  libraryId: string;
  name: string;
  category: string;
  description: string;
  content: string;
  variables: string[];
  qualityScore: number;
  usageCount: number;
  currentVersion: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface TemplateLibrary {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface Guide {
  id: string;
  title: string;
  templateId: string;
  content: string;
  variables: Record<string, string>;
  status: string;
  riskLevel: string;
  currentVersion: number;
  createdAt: string;
  updatedAt: string;
}

export interface Review {
  id: string;
  guideId: string;
  reviewer: string;
  status: string;
  decisionNote: string;
  createdAt: string;
  finishedAt: string;
}

export type UserRole = 'admin' | 'readonly' | 'developer' | 'ops' | 'tester' | 'architect' | 'designer';

export interface User {
  id: string;
  username: string;
  role: UserRole;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface LoginResult {
  token?: string;
  expiresAt?: string;
  user: User;
}

export interface PrecheckFinding {
  severity: string;
  category: string;
  excerpt: string;
  problem: string;
  suggestion: string;
}

export interface ImageInput {
  dataUrl?: string;
  url?: string;
  mimeType?: string;
}

export interface AgentSettings {
  provider: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  embeddingModel: string;
  timeoutSeconds: number;
  systemPrompt: string;
}

export interface AgentType {
  type: string;
  name: string;
  description: string;
}

export const api = {
  register: (username: string, password: string) =>
    http.post<LoginResult>('/auth/register', { username, password }).then((r) => r.data),
  login: (username: string, password: string) =>
    http.post<LoginResult>('/auth/login', { username, password }).then((r) => r.data),
  logout: () => http.post('/auth/logout'),
  me: () => http.get<{ user: User }>('/auth/me').then((r) => r.data.user),

  listTemplateLibraries: () => http.get<{ data: TemplateLibrary[] }>('/template-libraries').then((r) => r.data.data),
  createTemplateLibrary: (l: Partial<TemplateLibrary>) =>
    http.post<{ data: TemplateLibrary }>('/template-libraries', l).then((r) => r.data.data),
  listTemplates: (libraryId?: string) =>
    http.get<{ data: Template[] }>('/templates', { params: libraryId ? { libraryId } : {} }).then((r) => r.data.data),
  getTemplate: (id: string) => http.get<{ data: Template }>(`/templates/${id}`).then((r) => r.data.data),
  createTemplate: (t: Partial<Template>) => http.post<{ data: Template }>('/templates', t).then((r) => r.data.data),
  updateTemplate: (id: string, t: Partial<Template>) =>
    http.put<{ data: Template }>(`/templates/${id}`, t).then((r) => r.data.data),

  listGuides: () => http.get<{ data: Guide[] }>('/guides').then((r) => r.data.data),
  getGuide: (id: string) => http.get<{ data: Guide }>(`/guides/${id}`).then((r) => r.data.data),
  createGuide: (g: Partial<Guide>) => http.post<{ data: Guide }>('/guides', g).then((r) => r.data.data),
  updateGuide: (id: string, g: Partial<Guide>) =>
    http.put<{ data: Guide }>(`/guides/${id}`, g).then((r) => r.data.data),
  precheck: (content: string, images: ImageInput[] = []) =>
    http.post<{ data: { summary: string; findings: PrecheckFinding[] } }>('/guides/precheck', { content, images }).then((r) => r.data.data),

  listReviews: (guideId: string) =>
    http.get<{ data: Review[] }>(`/guides/${guideId}/reviews`).then((r) => r.data.data),
  createReview: (guideId: string, reviewer: string) =>
    http.post<{ data: Review }>(`/guides/${guideId}/reviews`, { reviewer }).then((r) => r.data.data),
  decideReview: (rid: string, status: string, note: string) =>
    http.post<{ data: Review }>(`/reviews/${rid}/decision`, { status, note }).then((r) => r.data.data),

  metrics: () => http.get<{ data: { issueCount: number; ruleCount: number } }>('/metrics/quality').then((r) => r.data.data),

  listUsers: () => http.get<{ data: User[] }>('/users').then((r) => r.data.data),
  createUser: (u: Partial<User>) => http.post<{ data: User }>('/users', u).then((r) => r.data.data),
  updateUser: (id: string, u: Partial<User>) => http.put<{ data: User }>(`/users/${id}`, u).then((r) => r.data.data),
  deleteUser: (id: string) => http.delete(`/users/${id}`),

  getAgentSettings: () => http.get<{ data: AgentSettings }>('/settings/agent').then((r) => r.data.data),
  updateAgentSettings: (s: AgentSettings) => http.put<{ data: AgentSettings }>('/settings/agent', s).then((r) => r.data.data),
  getAgentTypes: () => http.get<{ data: AgentType[] }>('/agent/types').then((r) => r.data.data),
  checkAgentHealth: () => http.post<{ data: { healthy: boolean; message?: string; error?: string } }>('/agent/health').then((r) => r.data.data),
};

// SSE 流式生成
export async function generateGuide(
  body: Record<string, unknown>,
  onChunk: (text: string) => void,
  onDone: () => void,
  onError: (msg: string) => void
) {
  const resp = await fetch('/api/guides/generate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!resp.body) {
    onError('无响应流');
    return;
  }
  const reader = resp.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split('\n\n');
    buffer = events.pop() || '';
    for (const evt of events) {
      const lines = evt.split('\n');
      let event = 'message';
      const dataLines: string[] = [];
      for (const line of lines) {
        if (line.startsWith('event:')) event = line.slice(6).trim();
        else if (line.startsWith('data:')) dataLines.push(line.slice(5));
      }
      const data = dataLines.join('\n');
      if (event === 'chunk') onChunk(data);
      else if (event === 'done') onDone();
      else if (event === 'error') onError(data);
    }
  }
  onDone();
}
