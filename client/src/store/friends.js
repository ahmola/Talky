import { defineStore } from 'pinia';
import { useAuthStore } from './auth';
import { getLocalFriends, saveLocalFriend, clearFriends, saveLocalUser } from '../db/sqlite';

export const useFriendsStore = defineStore('friends', {
  state: () => ({
    friends: [],
  }),
  actions: {
    async syncFriends() {
      const authStore = useAuthStore();
      const currentUserID = authStore.userId;
      if (!currentUserID) return;

      // 1. 로컬 캐시 우선 로드
      try {
        if (authStore.dbReady) {
          this.friends = getLocalFriends(currentUserID);
        }
      } catch (err) {
        console.error('Local friends fetch failed:', err);
      }

      // 2. 서버 데이터와 병합 (동기화)
      try {
        const res = await fetch('/api/friends', {
          headers: { 'Authorization': `Bearer ${authStore.token}` }
        });
        if (res.ok) {
          const serverFriends = await res.json();
          
          if (authStore.dbReady) {
            clearFriends(currentUserID);
            serverFriends.forEach(f => {
              saveLocalFriend(currentUserID, f);
            });
            this.friends = getLocalFriends(currentUserID);
          } else {
            this.friends = serverFriends;
          }
        }
      } catch (err) {
        console.error('Server friends sync failed:', err);
      }
    },
    async addFriend(username) {
      const authStore = useAuthStore();
      const res = await fetch('/api/friends', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${authStore.token}`
        },
        body: JSON.stringify({ username }),
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Friend add failed');
      }
      await this.syncFriends();
    },
    async deleteFriend(friendId) {
      const authStore = useAuthStore();
      const res = await fetch(`/api/friends/${friendId}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${authStore.token}` }
      });
      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.error || 'Friend delete failed');
      }
      await this.syncFriends();
    },
    handleStatusChange(data) {
      const { userId, online } = data;
      const friend = this.friends.find(f => f.userId === userId);
      if (friend) {
        friend.online = online;
        
        const authStore = useAuthStore();
        if (authStore.dbReady) {
          try {
            saveLocalUser({
              id: userId,
              username: friend.username,
              nickname: friend.nickname,
              statusMessage: friend.statusMessage,
              online: online
            });
          } catch (err) {
            console.error('Failed to update local user status:', err);
          }
        }
      }
    }
  }
});
