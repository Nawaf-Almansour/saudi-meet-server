const API_BASE = '/api';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('diwan_token');
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...((options.headers as Record<string, string>) || {}),
  };

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export const api = {
  // Auth
  register: (email: string, password: string, display_name: string) =>
    request<{ token: string; user: any }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, display_name }),
    }),

  login: (email: string, password: string) =>
    request<{ token: string; user: any }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  me: () => request<any>('/auth/me'),

  // Workspaces
  createWorkspace: (name: string) =>
    request<any>('/workspaces', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  listWorkspaces: () => request<any[]>('/workspaces'),

  discoverWorkspaces: () => request<any[]>('/workspaces/discover'),

  joinWorkspace: (slug: string) =>
    request<any>('/workspaces/join', {
      method: 'POST',
      body: JSON.stringify({ slug }),
    }),

  // Channels
  listChannels: (workspaceId: string) =>
    request<any[]>(`/workspaces/${workspaceId}/channels`),

  createChannel: (workspaceId: string, name: string, description?: string) =>
    request<any>(`/workspaces/${workspaceId}/channels`, {
      method: 'POST',
      body: JSON.stringify({ name, description, visibility: 'public' }),
    }),

  joinChannel: (channelId: string) =>
    request<any>(`/channels/${channelId}/join`, { method: 'POST' }),

  // Messages
  sendMessage: (channelId: string, body: string) =>
    request<any>(`/channels/${channelId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ body }),
    }),

  listMessages: (channelId: string, limit = 50, offset = 0) =>
    request<any[]>(`/channels/${channelId}/messages?limit=${limit}&offset=${offset}`),

  // Meetings
  startMeeting: (channelId: string, title?: string) =>
    request<{ room_id: string; join_url: string; title: string }>(`/channels/${channelId}/meetings/start`, {
      method: 'POST',
      body: JSON.stringify({ title: title || '' }),
    }),

  joinMeeting: (channelId: string) =>
    request<{ room_id: string; join_url: string }>(`/channels/${channelId}/meetings/join`, {
      method: 'POST',
    }),
};
