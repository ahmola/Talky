<template>
  <div class="friends-container">
    <div class="header">
      <h2>친구 ({{ friendsStore.friends.length }})</h2>
      <button class="add-friend-btn" @click="showAddModal = true" title="친구 추가">👤+</button>
    </div>

    <!-- 검색 바 -->
    <div class="search-box">
      <input v-model="searchQuery" type="text" placeholder="친구 검색..." class="search-input" />
    </div>

    <!-- 친구 리스트 -->
    <div class="friends-list">
      <div v-if="filteredFriends.length === 0" class="empty-list">
        {{ searchQuery ? '검색 결과가 없습니다.' : '등록된 친구가 없습니다.' }}
      </div>
      <div 
        v-for="friend in filteredFriends" 
        :key="friend.userId" 
        class="friend-item"
        @click="startDirectChat(friend)"
      >
        <div class="avatar-container">
          <div class="avatar">{{ friend.nickname ? friend.nickname[0] : 'F' }}</div>
          <div class="status-dot" :class="{ online: friend.online }"></div>
        </div>
        <div class="friend-info">
          <div class="nickname">{{ friend.nickname }}</div>
          <div class="status-msg">{{ friend.statusMessage || '상태 메시지가 없습니다.' }}</div>
        </div>
        <!-- 삭제 버튼 (마우스 오버 시 표시) -->
        <button class="delete-btn" @click.stop="handleDeleteFriend(friend.userId)" title="친구 삭제">✕</button>
      </div>
    </div>

    <!-- 친구 추가 모달 -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
      <div class="modal-content">
        <h3>친구 추가</h3>
        <p>추가하려는 친구의 아이디(Username)를 입력해주세요.</p>
        <input v-model="newFriendUsername" type="text" class="form-input" placeholder="예: user123" @keyup.enter="handleAddFriend" />
        <div class="modal-actions">
          <button class="btn-secondary" @click="showAddModal = false">취소</button>
          <button class="btn-primary" @click="handleAddFriend">추가</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useFriendsStore } from '../store/friends';
import { useChatStore } from '../store/chat';
import { useAuthStore } from '../store/auth';

const friendsStore = useFriendsStore();
const chatStore = useChatStore();
const authStore = useAuthStore();

const searchQuery = ref('');
const showAddModal = ref(false);
const newFriendUsername = ref('');

const filteredFriends = computed(() => {
  const query = searchQuery.value.toLowerCase().trim();
  if (!query) return friendsStore.friends;
  return friendsStore.friends.filter(f => 
    f.nickname.toLowerCase().includes(query) || 
    f.username.toLowerCase().includes(query)
  );
});

const handleAddFriend = async () => {
  if (!newFriendUsername.value.trim()) return;
  try {
    await friendsStore.addFriend(newFriendUsername.value.trim());
    alert('친구를 추가했습니다.');
    newFriendUsername.value = '';
    showAddModal.value = false;
  } catch (err) {
    alert(err.message);
  }
};

const handleDeleteFriend = async (friendId) => {
  if (!confirm('이 친구를 삭제하시겠습니까?')) return;
  try {
    await friendsStore.deleteFriend(friendId);
  } catch (err) {
    alert(err.message);
  }
};

const startDirectChat = async (friend) => {
  let targetRoomId = null;
  
  // 1:1 대화방이 이미 존재하는지 검사
  const existingRoom = chatStore.rooms.find(r => r.type === 'direct' && r.name.includes(friend.nickname));
  
  if (existingRoom) {
    targetRoomId = existingRoom.roomId;
  } else {
    try {
      targetRoomId = await chatStore.createRoom(
        `${friend.nickname}, ${authStore.nickname}`,
        'direct',
        [friend.userId]
      );
    } catch (err) {
      alert('대화방 생성 실패: ' + err.message);
      return;
    }
  }

  chatStore.activeRoomId = targetRoomId;
  chatStore.loadMessages(targetRoomId);
};
</script>

<style scoped>
.friends-container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 48px;
}

.header h2 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-header);
}

.add-friend-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: var(--text-normal);
  transition: var(--transition-smooth);
}

.add-friend-btn:hover {
  color: var(--accent-blurple);
  transform: scale(1.1);
}

.search-box {
  padding: 8px 12px;
}

.search-input {
  width: 100%;
  padding: 6px 10px;
  background-color: var(--bg-tertiary);
  border: none;
  border-radius: 4px;
  outline: none;
  font-size: 14px;
}

.friends-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.empty-list {
  text-align: center;
  padding: 24px;
  color: var(--text-muted);
  font-size: 14px;
}

.friend-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: var(--transition-smooth);
  position: relative;
  gap: 12px;
}

.friend-item:hover {
  background-color: var(--bg-modifier-hover);
}

.avatar-container {
  position: relative;
  width: 36px;
  height: 36px;
}

.avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background-color: var(--accent-blurple);
  display: flex;
  justify-content: center;
  align-items: center;
  font-weight: 600;
  font-size: 15px;
  color: white;
}

.status-dot {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background-color: var(--text-muted);
  border: 2px solid var(--bg-secondary);
  transition: var(--transition-smooth);
}

.status-dot.online {
  background-color: var(--accent-green);
  box-shadow: 0 0 4px var(--accent-green);
}

.friend-info {
  flex: 1;
  min-width: 0;
}

.nickname {
  color: var(--text-normal);
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-msg {
  color: var(--text-muted);
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.delete-btn {
  display: none;
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 14px;
}

.friend-item:hover .delete-btn {
  display: block;
}

.delete-btn:hover {
  color: var(--accent-red);
}
</style>
