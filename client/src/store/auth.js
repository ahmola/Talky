import { defineStore } from 'pinia';
import { initDB } from '../db/sqlite';
import { useFriendsStore } from './friends';
import { useChatStore } from './chat';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userId: localStorage.getItem('userId') || '',
    nickname: localStorage.getItem('nickname') || '',
    ws: null,
    dbReady: false,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
  },
  actions: {
    async initialize() {
      if (this.isAuthenticated) {
        try {
          await initDB();
          this.dbReady = true;
          this.connectWebSocket();
          
          // 백그라운드 데이터 동기화
          useFriendsStore().syncFriends();
          useChatStore().syncRooms();
        } catch (err) {
          console.error('SQLite initialization failed:', err);
        }
      }
    },
    async register(username, password, nickname) {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: username.trim(),
          password,
          nickname: nickname.trim(),
        }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Registration failed');
      }
      return res.json();
    },
    async login(username, password) {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          username: username.trim(),
          password,
        }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Login failed');
      }
      const data = await res.json();
      this.token = data.token;
      this.userId = data.userId;
      this.nickname = data.nickname;
      localStorage.setItem('token', data.token);
      localStorage.setItem('userId', data.userId);
      localStorage.setItem('nickname', data.nickname);

      await initDB();
      this.dbReady = true;
      this.connectWebSocket();

      // 로그인 성공 시 동기화 기동
      useFriendsStore().syncFriends();
      useChatStore().syncRooms();
    },
    logout() {
      this.token = '';
      this.userId = '';
      this.nickname = '';
      localStorage.removeItem('token');
      localStorage.removeItem('userId');
      localStorage.removeItem('nickname');
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }
    },
    connectWebSocket() {
      if (this.ws) return;
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.host;
      const wsUrl = `${protocol}//${host}/ws?token=${this.token}`;
      
      this.ws = new WebSocket(wsUrl);

      this.ws.onmessage = (event) => {
        try {
          const wsMsg = JSON.parse(event.data);
          this.handleWSEvent(wsMsg.event, wsMsg.data);
        } catch (e) {
          console.error('WS message parse error:', e);
        }
      };

      this.ws.onclose = () => {
        this.ws = null;
        // 의도치 않은 접속 해제 시 3초 후 자동 재시도
        if (this.isAuthenticated) {
          setTimeout(() => this.connectWebSocket(), 3000);
        }
      };
    },
    sendMessage(roomId, content) {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        throw new Error('WebSocket is not open');
      }
      this.ws.send(JSON.stringify({
        event: 'send_message',
        data: { roomId, content }
      }));
    },
    handleWSEvent(event, data) {
      const friendsStore = useFriendsStore();
      const chatStore = useChatStore();

      switch (event) {
        case 'new_message':
          chatStore.handleNewMessage(data);
          break;
        case 'status_change':
          friendsStore.handleStatusChange(data);
          break;
        case 'error':
          console.error('WS Error:', data.message);
          break;
      }
    }
  }
});
