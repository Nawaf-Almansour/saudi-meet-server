import { create } from 'zustand';

interface User {
  id: string;
  email: string;
  display_name: string;
  avatar_url: string;
  status: string;
}

interface Workspace {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
}

interface Channel {
  id: string;
  workspace_id: string;
  name: string;
  slug: string;
  description: string;
}

interface Message {
  id: string;
  channel_id: string;
  sender_id: string;
  sender_name: string;
  body: string;
  message_type: string;
  created_at: string;
}

interface AppState {
  user: User | null;
  token: string | null;
  workspaces: Workspace[];
  currentWorkspace: Workspace | null;
  channels: Channel[];
  currentChannel: Channel | null;
  messages: Message[];
  wsConnection: WebSocket | null;

  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  setWorkspaces: (workspaces: Workspace[]) => void;
  setCurrentWorkspace: (workspace: Workspace | null) => void;
  setChannels: (channels: Channel[]) => void;
  setCurrentChannel: (channel: Channel | null) => void;
  setMessages: (messages: Message[]) => void;
  addMessage: (message: Message) => void;
  setWsConnection: (ws: WebSocket | null) => void;
  logout: () => void;
}

export const useStore = create<AppState>((set) => ({
  user: null,
  token: localStorage.getItem('diwan_token'),
  workspaces: [],
  currentWorkspace: null,
  channels: [],
  currentChannel: null,
  messages: [],
  wsConnection: null,

  setUser: (user) => set({ user }),
  setToken: (token) => {
    if (token) {
      localStorage.setItem('diwan_token', token);
    } else {
      localStorage.removeItem('diwan_token');
    }
    set({ token });
  },
  setWorkspaces: (workspaces) => set({ workspaces }),
  setCurrentWorkspace: (workspace) => set({ currentWorkspace: workspace }),
  setChannels: (channels) => set({ channels }),
  setCurrentChannel: (channel) => set({ currentChannel: channel }),
  setMessages: (messages) => set({ messages }),
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  setWsConnection: (ws) => set({ wsConnection: ws }),
  logout: () => {
    localStorage.removeItem('diwan_token');
    set({ user: null, token: null, workspaces: [], currentWorkspace: null, channels: [], currentChannel: null, messages: [], wsConnection: null });
  },
}));
