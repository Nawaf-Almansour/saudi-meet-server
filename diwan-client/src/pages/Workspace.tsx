import { useEffect, useState, useRef } from 'react';
import { Hash, Plus, Send, LogOut, MessageSquare, Video, ExternalLink } from 'lucide-react';
import { api } from '../api/client';
import { useStore } from '../store';

export default function Workspace() {
  const {
    user, token, workspaces, currentWorkspace, channels,
    currentChannel, messages, setUser, setWorkspaces,
    setCurrentWorkspace, setChannels, setCurrentChannel,
    setMessages, addMessage, logout,
  } = useStore();

  const [msgInput, setMsgInput] = useState('');
  const [newChannelName, setNewChannelName] = useState('');
  const [showNewChannel, setShowNewChannel] = useState(false);
  const [newWorkspaceName, setNewWorkspaceName] = useState('');
  const [showNewWorkspace, setShowNewWorkspace] = useState(false);
  const [joinSlug, setJoinSlug] = useState('');
  const [showJoinWorkspace, setShowJoinWorkspace] = useState(false);
  const [availableWorkspaces, setAvailableWorkspaces] = useState<any[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // Load user info
  useEffect(() => {
    if (token && !user) {
      api.me().then(setUser).catch(() => logout());
    }
  }, [token]);

  // Load workspaces
  useEffect(() => {
    if (token) {
      api.listWorkspaces().then((ws) => {
        setWorkspaces(ws);
        if (ws.length > 0 && !currentWorkspace) {
          setCurrentWorkspace(ws[0]);
        }
      });
    }
  }, [token]);

  // Load channels when workspace changes
  useEffect(() => {
    if (currentWorkspace) {
      api.listChannels(currentWorkspace.id).then((chs) => {
        setChannels(chs);
        if (chs.length > 0 && !currentChannel) {
          setCurrentChannel(chs[0]);
        }
      });
    }
  }, [currentWorkspace]);

  // Load messages when channel changes
  useEffect(() => {
    if (currentChannel) {
      api.listMessages(currentChannel.id).then(setMessages);
    }
  }, [currentChannel]);

  // WebSocket connection
  useEffect(() => {
    if (!token) return;

    const wsUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/api/ws?token=${token}`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      if (currentChannel) {
        ws.send(JSON.stringify({ action: 'subscribe', channel_id: currentChannel.id }));
      }
    };

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (data.type === 'message.new') {
        addMessage(data.payload);
      }
    };

    return () => {
      ws.close();
    };
  }, [token]);

  // Subscribe to channel via WS
  useEffect(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN && currentChannel) {
      wsRef.current.send(JSON.stringify({ action: 'subscribe', channel_id: currentChannel.id }));
    }
  }, [currentChannel]);

  // Scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!msgInput.trim() || !currentChannel) return;
    await api.sendMessage(currentChannel.id, msgInput.trim());
    setMsgInput('');
  };

  const handleStartMeeting = async () => {
    if (!currentChannel) return;
    try {
      const res = await api.startMeeting(currentChannel.id);
      window.open(res.join_url, '_blank');
    } catch (err: any) {
      alert(err.message || 'فشل بدء الاجتماع');
    }
  };

  const handleJoinMeeting = async () => {
    if (!currentChannel) return;
    try {
      const res = await api.joinMeeting(currentChannel.id);
      window.open(res.join_url, '_blank');
    } catch (err: any) {
      alert(err.message || 'فشل الانضمام للاجتماع');
    }
  };

  const handleCreateWorkspace = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newWorkspaceName.trim()) return;
    const ws = await api.createWorkspace(newWorkspaceName.trim());
    setWorkspaces([...workspaces, ws]);
    setCurrentWorkspace(ws);
    setNewWorkspaceName('');
    setShowNewWorkspace(false);
  };

  const handleShowJoinWorkspace = async () => {
    setShowJoinWorkspace(true);
    try {
      const all = await api.discoverWorkspaces();
      const myIds = workspaces.map((w: any) => w.id);
      setAvailableWorkspaces(all.filter((w: any) => !myIds.includes(w.id)));
    } catch {
      setAvailableWorkspaces([]);
    }
  };

  const handleJoinWorkspaceBySlug = async (slug: string) => {
    try {
      const res = await api.joinWorkspace(slug);
      const ws = res.workspace;
      setWorkspaces([...workspaces, ws]);
      setCurrentWorkspace(ws);
      setShowJoinWorkspace(false);
      setJoinSlug('');
    } catch (err: any) {
      alert(err.message || 'فشل الانضمام');
    }
  };

  const handleCreateChannel = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newChannelName.trim() || !currentWorkspace) return;
    const ch = await api.createChannel(currentWorkspace.id, newChannelName.trim());
    setChannels([...channels, ch]);
    setCurrentChannel(ch);
    setNewChannelName('');
    setShowNewChannel(false);
  };

  return (
    <div className="h-screen flex">
      {/* Sidebar */}
      <aside className="w-64 bg-sidebar text-sidebar-text flex flex-col">
        {/* Workspace Header */}
        <div className="p-4 border-b border-sidebar-light">
          <div className="flex items-center justify-between">
            <h2 className="text-white font-bold text-lg truncate">
              {currentWorkspace?.name || 'ديوان'}
            </h2>
            <button onClick={logout} className="text-sidebar-text hover:text-white" title="خروج">
              <LogOut size={18} />
            </button>
          </div>
          <p className="text-xs mt-1 opacity-70">{user?.display_name}</p>
        </div>

        {/* Channels */}
        <div className="flex-1 overflow-y-auto p-3">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-semibold uppercase tracking-wide">القنوات</span>
            <button
              onClick={() => setShowNewChannel(true)}
              className="text-sidebar-text hover:text-white"
            >
              <Plus size={16} />
            </button>
          </div>

          {showNewChannel && (
            <form onSubmit={handleCreateChannel} className="mb-2">
              <input
                type="text"
                value={newChannelName}
                onChange={(e) => setNewChannelName(e.target.value)}
                placeholder="اسم القناة"
                className="w-full px-2 py-1 rounded text-sm bg-sidebar-light text-white placeholder-sidebar-text outline-none"
                autoFocus
              />
            </form>
          )}

          {channels.map((ch) => (
            <button
              key={ch.id}
              onClick={() => setCurrentChannel(ch)}
              className={`w-full text-right px-3 py-1.5 rounded text-sm flex items-center gap-2 mb-0.5 ${
                currentChannel?.id === ch.id
                  ? 'bg-sidebar-active text-white'
                  : 'hover:bg-sidebar-light'
              }`}
            >
              <Hash size={14} />
              <span className="truncate">{ch.name}</span>
            </button>
          ))}
        </div>

        {/* Workspace List */}
        <div className="p-3 border-t border-sidebar-light">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-semibold uppercase tracking-wide">المساحات</span>
            <button
              onClick={() => setShowNewWorkspace(true)}
              className="text-sidebar-text hover:text-white"
            >
              <Plus size={16} />
            </button>
          </div>

          {showNewWorkspace && (
            <form onSubmit={handleCreateWorkspace} className="mb-2">
              <input
                type="text"
                value={newWorkspaceName}
                onChange={(e) => setNewWorkspaceName(e.target.value)}
                placeholder="اسم المساحة"
                className="w-full px-2 py-1 rounded text-sm bg-sidebar-light text-white placeholder-sidebar-text outline-none"
                autoFocus
              />
            </form>
          )}

          {workspaces.map((ws) => (
            <button
              key={ws.id}
              onClick={() => {
                setCurrentWorkspace(ws);
                setCurrentChannel(null);
                setChannels([]);
                setMessages([]);
              }}
              className={`w-full text-right px-2 py-1 rounded text-xs mb-0.5 ${
                currentWorkspace?.id === ws.id ? 'bg-sidebar-active text-white' : 'hover:bg-sidebar-light'
              }`}
            >
              {ws.name}
            </button>
          ))}

          <button
            onClick={handleShowJoinWorkspace}
            className="w-full text-right px-2 py-1.5 mt-2 rounded text-xs text-sidebar-text hover:bg-sidebar-light border border-dashed border-sidebar-text/30"
          >
            + انضم لمساحة عمل
          </button>

          {showJoinWorkspace && (
            <div className="mt-2 space-y-1">
              <input
                type="text"
                value={joinSlug}
                onChange={(e) => setJoinSlug(e.target.value)}
                placeholder="أدخل رمز المساحة..."
                className="w-full px-2 py-1 rounded text-sm bg-sidebar-light text-white placeholder-sidebar-text outline-none"
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && joinSlug.trim()) {
                    handleJoinWorkspaceBySlug(joinSlug.trim());
                  }
                }}
              />
              {availableWorkspaces.length > 0 && (
                <div className="space-y-0.5">
                  <span className="text-[10px] text-sidebar-text/60 uppercase">مساحات متاحة</span>
                  {availableWorkspaces.map((aw: any) => (
                    <button
                      key={aw.id}
                      onClick={() => handleJoinWorkspaceBySlug(aw.slug)}
                      className="w-full text-right px-2 py-1 rounded text-xs hover:bg-sidebar-light text-green-300"
                    >
                      + {aw.name}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </aside>

      {/* Main Chat Area */}
      <main className="flex-1 flex flex-col bg-white">
        {/* Channel Header */}
        {currentChannel ? (
          <>
            <header className="h-14 border-b flex items-center justify-between px-5">
              <div className="flex items-center gap-2">
                <Hash size={18} className="text-gray-500" />
                <h3 className="font-semibold text-gray-900">{currentChannel.name}</h3>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={handleJoinMeeting}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-green-700 bg-green-50 hover:bg-green-100 rounded-lg transition-colors"
                  title="انضم للاجتماع"
                >
                  <ExternalLink size={14} />
                  <span>انضمام</span>
                </button>
                <button
                  onClick={handleStartMeeting}
                  className="flex items-center gap-1.5 px-3 py-1.5 text-sm text-white bg-purple-700 hover:bg-purple-800 rounded-lg transition-colors"
                  title="بدء اجتماع"
                >
                  <Video size={14} />
                  <span>اجتماع</span>
                </button>
              </div>
            </header>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-5 space-y-4">
              {messages.map((msg) => (
                <div key={msg.id} className="flex gap-3">
                  <div className="w-9 h-9 rounded-lg bg-purple-100 flex items-center justify-center text-purple-700 font-bold text-sm shrink-0">
                    {msg.sender_name?.charAt(0) || '؟'}
                  </div>
                  <div>
                    <div className="flex items-baseline gap-2">
                      <span className="font-semibold text-sm text-gray-900">{msg.sender_name}</span>
                      <span className="text-xs text-gray-400">
                        {new Date(msg.created_at).toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })}
                      </span>
                    </div>
                    <p className="text-gray-700 text-sm mt-0.5">{msg.body}</p>
                  </div>
                </div>
              ))}
              <div ref={messagesEndRef} />
            </div>

            {/* Message Composer */}
            <form onSubmit={handleSend} className="p-4 border-t">
              <div className="flex items-center gap-3 bg-gray-50 rounded-xl px-4 py-2 border">
                <input
                  type="text"
                  value={msgInput}
                  onChange={(e) => setMsgInput(e.target.value)}
                  placeholder={`رسالة في #${currentChannel.name}`}
                  className="flex-1 bg-transparent outline-none text-sm"
                />
                <button
                  type="submit"
                  disabled={!msgInput.trim()}
                  className="text-purple-700 hover:text-purple-900 disabled:text-gray-300 transition-colors"
                >
                  <Send size={20} />
                </button>
              </div>
            </form>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-gray-400">
            <div className="text-center">
              <MessageSquare size={48} className="mx-auto mb-4 opacity-50" />
              <p className="text-lg">اختر قناة لبدء المحادثة</p>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
