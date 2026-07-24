<template>
  <div class="chatlist-container">
    <div class="header">
      <h2>채팅 ({{ chatStore.rooms.length }})</h2>
      <button class="create-room-btn" @click="openCreateModal" title="새로운 채팅방">+💬</button>
    </div>

    <!-- 검색 바 -->
    <div class="search-box">
      <input v-model="searchQuery" type="text" placeholder="채팅방 검색..." class="search-input" />
    </div>

    <!-- 대화방 리스트 -->
    <div class="rooms-list">
      <div v-if="filteredRooms.length === 0" class="empty-list">
        {{ searchQuery ? '검색 결과가 없습니다.' : '대화방이 없습니다.' }}
      </div>
      <div 
        v-for="room in filteredRooms" 
        :key="room.roomId" 
        class="room-item"
        :class="{ active: chatStore.activeRoomId === room.roomId }"
        @click="openRoom(room.roomId)"
      >
        <div class="room-avatar">
          {{ room.type === 'direct' ? '👤' : '👥' }}
        </div>
        <div class="room-info">
          <div class="room-name">{{ getRoomName(room) }}</div>
          <div class="room-type-badge" :class="room.type">{{ room.type === 'direct' ? '1:1' : '그룹' }}</div>
        </div>
      </div>
    </div>

    <!-- 대화방 생성 모달 -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal-content">
        <h3>대화방 개설</h3>
        
        <div class="form-group">
          <label>대화방 이름</label>
          <input v-model="newRoomName" type="text" class="form-input" placeholder="예: 코딩 대화방" />
        </div>

        <div class="form-group">
          <label>초대할 친구 선택</label>
          <div class="invite-friends-list">
            <div v-if="friendsStore.friends.length === 0" class="no-friends">
              초대 가능한 친구가 없습니다.
            </div>
            <div v-for="friend in friendsStore.friends" :key="friend.userId" class="invite-item">
              <label class="checkbox-container">
                <input type="checkbox" :value="friend.userId" v-model="selectedFriendIDs" />
                {{ friend.nickname }} (@{{ friend.username }})
              </label>
            </div>
          </div>
        </div>

        <div class="modal-actions">
          <button class="btn-secondary" @click="showCreateModal = false">취소</button>
          <button class="btn-primary" @click="handleCreateRoom">생성</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useChatStore } from '../store/chat';
import { useFriendsStore } from '../store/friends';
import { useAuthStore } from '../store/auth';

const chatStore = useChatStore();
const friendsStore = useFriendsStore();
const authStore = useAuthStore();

const searchQuery = ref('');
const showCreateModal = ref(false);
const newRoomName = ref('');
const selectedFriendIDs = ref([]);

const filteredRooms = computed(() => {
  const query = searchQuery.value.toLowerCase().trim();
  if (!query) return chatStore.rooms;
  return chatStore.rooms.filter(r => 
    r.name.toLowerCase().includes(query)
  );
});

const getRoomName = (room) => {
  if (room.type === 'direct') {
    const parts = room.name.split(', ');
    if (parts.length > 1) {
      return parts[0] === authStore.nickname ? parts[1] : parts[0];
    }
  }
  return room.name;
};

const openRoom = (roomId) => {
  chatStore.activeRoomId = roomId;
  chatStore.loadMessages(roomId);
};

const openCreateModal = () => {
  newRoomName.value = '';
  selectedFriendIDs.value = [];
  showCreateModal.value = true;
};

const handleCreateRoom = async () => {
  if (selectedFriendIDs.value.length === 0) {
    alert('초대할 친구를 한 명 이상 선택해주세요.');
    return;
  }

  const roomType = selectedFriendIDs.value.length === 1 ? 'direct' : 'group';
  let finalName = newRoomName.value.trim();
  if (!finalName) {
    const selectedNicknames = friendsStore.friends
      .filter(f => selectedFriendIDs.value.includes(f.userId))
      .map(f => f.nickname);
    finalName = selectedNicknames.join(', ');
  }

  try {
    const roomId = await chatStore.createRoom(finalName, roomType, selectedFriendIDs.value);
    showCreateModal.value = false;
    openRoom(roomId);
  } catch (err) {
    alert(err.message);
  }
};
</script>

<style scoped>
.chatlist-container {
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

.create-room-btn {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: var(--text-normal);
  transition: var(--transition-smooth);
}

.create-room-btn:hover {
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

.rooms-list {
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

.room-item {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: var(--transition-smooth);
  gap: 12px;
}

.room-item:hover {
  background-color: var(--bg-modifier-hover);
}

.room-item.active {
  background-color: var(--bg-modifier-selected);
  color: var(--text-header);
}

.room-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background-color: var(--bg-tertiary);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 20px;
}

.room-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.room-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-normal);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.room-type-badge {
  font-size: 10px;
  align-self: flex-start;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 600;
  text-transform: uppercase;
}

.room-type-badge.direct {
  background-color: rgba(88, 101, 242, 0.2);
  color: var(--accent-blurple);
}

.room-type-badge.group {
  background-color: rgba(35, 165, 90, 0.2);
  color: var(--accent-green);
}

.invite-friends-list {
  max-height: 180px;
  overflow-y: auto;
  background-color: var(--bg-tertiary);
  padding: 8px;
  border-radius: 4px;
  border: 1px solid rgba(0,0,0,0.3);
}

.no-friends {
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
  padding: 12px;
}

.invite-item {
  padding: 6px 4px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
}

.checkbox-container {
  display: flex;
  align-items: center;
  cursor: pointer;
  font-size: 14px;
  user-select: none;
  gap: 8px;
}
</style>
