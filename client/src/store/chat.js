import { defineStore } from 'pinia';
import { useAuthStore } from './auth';
import { getLocalRooms, saveLocalRoom, getLocalMessages, saveLocalMessage } from '../db/sqlite';

export const useChatStore = defineStore('chat', {
  state: () => ({
    rooms: [],
    messages: {}, // {[roomId]: Message[]}
    activeRoomId: null,
  }),
  actions: {
    async syncRooms() {
      const authStore = useAuthStore();
      if (!authStore.userId) return;

      // 1. 로컬 SQLite 우선 조회
      try {
        if (authStore.dbReady) {
          this.rooms = getLocalRooms();
        }
      } catch (err) {
        console.error('Failed to fetch local rooms:', err);
      }

      // 2. 서버와 병합
      try {
        const res = await fetch('/api/rooms', {
          headers: { 'Authorization': `Bearer ${authStore.token}` }
        });
        if (res.ok) {
          const serverRooms = await res.json();
          if (authStore.dbReady) {
            serverRooms.forEach(r => {
              saveLocalRoom(r);
            });
            this.rooms = getLocalRooms();
          } else {
            this.rooms = serverRooms;
          }
        }
      } catch (err) {
        console.error('Failed to sync server rooms:', err);
      }
    },
    async createRoom(name, type, members) {
      const authStore = useAuthStore();
      const res = await fetch('/api/rooms', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({ name, type, members }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Room creation failed');
      }
      const data = await res.json();
      await this.syncRooms();
      return data.roomId;
    },
    async loadMessages(roomId) {
      const authStore = useAuthStore();
      
      // 1. 로컬 SQLite에서 로드
      if (authStore.dbReady) {
        this.messages[roomId] = getLocalMessages(roomId);
      } else {
        this.messages[roomId] = [];
      }

      // 2. 서버에서 메시지 백업 동기화 (최근 100개 일괄 캐싱)
      try {
        const res = await fetch(`/api/rooms/${roomId}/messages?limit=100`, {
          headers: { 'Authorization': `Bearer ${authStore.token}` }
        });
        if (res.ok) {
          const serverMessages = await res.json();
          if (authStore.dbReady) {
            serverMessages.forEach(msg => {
              saveLocalMessage(msg);
            });
            this.messages[roomId] = getLocalMessages(roomId);
          } else {
            this.messages[roomId] = serverMessages;
          }
        }
      } catch (err) {
        console.error('Failed to sync room messages:', err);
      }
    },
    async handleNewMessage(msg) {
      const authStore = useAuthStore();
      const { roomId } = msg;

      // 1. SQLite 캐시 저장
      if (authStore.dbReady) {
        try {
          saveLocalMessage(msg);
        } catch (err) {
          console.error('SQLite message caching failed:', err);
        }
      }

      // 2. 메모리 메시지 배열 최신화
      if (!this.messages[roomId]) {
        this.messages[roomId] = [];
      }
      
      const exists = this.messages[roomId].some(m => m.messageId === msg.messageId);
      if (!exists) {
        this.messages[roomId].push(msg);
      }

      // 3. 방 목록 updated_at 정렬 최신화
      await this.syncRooms();
    }
  }
});
