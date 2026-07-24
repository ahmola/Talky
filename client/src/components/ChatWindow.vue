<template>
  <div class="chat-window">
    <!-- 채팅방 헤더 -->
    <div class="chat-header">
      <span class="room-icon">💬</span>
      <h2 class="room-title">{{ roomName }}</h2>
    </div>

    <!-- 메시지 타임라인 -->
    <div class="messages-timeline" ref="timelineRef">
      <div v-if="!currentMessages || currentMessages.length === 0" class="no-messages">
        대화의 시작점입니다. 메시지를 전송해보세요!
      </div>
      <div 
        v-for="(msg, index) in currentMessages" 
        :key="msg.messageId" 
        class="message-wrapper"
        :class="{ 'same-sender': isSameSender(msg, index) }"
      >
        <!-- 프로필 아바타 (첫 메시지인 경우에만 렌더링) -->
        <div v-if="!isSameSender(msg, index)" class="message-avatar">
          {{ msg.senderNickname ? msg.senderNickname[0] : 'U' }}
        </div>
        <div v-else class="message-spacer"></div>

        <div class="message-content-area">
          <!-- 송신자 정보 (첫 메시지인 경우에만 렌더링) -->
          <div v-if="!isSameSender(msg, index)" class="message-meta">
            <span class="sender-name" :class="{ 'me': msg.senderId === authStore.userId }">
              {{ msg.senderNickname }}
            </span>
            <span class="send-time">{{ formatTime(msg.createdAt) }}</span>
          </div>

          <!-- 메시지 텍스트 본문 -->
          <div class="message-text">
            {{ msg.content }}
          </div>
        </div>
      </div>
    </div>

    <!-- 메시지 입력창 -->
    <div class="input-bar">
      <form @submit.prevent="handleSend" class="input-form">
        <input 
          v-model="inputContent" 
          type="text" 
          placeholder="메시지를 입력하세요..." 
          class="chat-input" 
          ref="inputRef"
        />
        <button type="submit" class="send-btn" :disabled="!inputContent.trim()">전송</button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue';
import { useChatStore } from '../store/chat';
import { useAuthStore } from '../store/auth';

const chatStore = useChatStore();
const authStore = useAuthStore();

const timelineRef = ref(null);
const inputRef = ref(null);
const inputContent = ref('');

const currentRoom = computed(() => {
  return chatStore.rooms.find(r => r.roomId === chatStore.activeRoomId);
});

const roomName = computed(() => {
  if (!currentRoom.value) return '';
  if (currentRoom.value.type === 'direct') {
    const parts = currentRoom.value.name.split(', ');
    if (parts.length > 1) {
      return parts[0] === authStore.nickname ? parts[1] : parts[0];
    }
  }
  return currentRoom.value.name;
});

const currentMessages = computed(() => {
  return chatStore.messages[chatStore.activeRoomId] || [];
});

const isSameSender = (msg, index) => {
  if (index === 0) return false;
  const prevMsg = currentMessages.value[index - 1];
  if (prevMsg.senderId !== msg.senderId) return false;
  
  // 1분 이내에 연속해서 같은 사람이 보낸 메시지는 프로필을 합쳐 그림
  const prevTime = new Date(prevMsg.createdAt).getTime();
  const currTime = new Date(msg.createdAt).getTime();
  return (currTime - prevTime) < 60000;
};

const handleSend = () => {
  if (!inputContent.value.trim()) return;
  try {
    authStore.sendMessage(chatStore.activeRoomId, inputContent.value.trim());
    inputContent.value = '';
    focusInput();
  } catch (err) {
    alert(err.message);
  }
};

const scrollToBottom = () => {
  nextTick(() => {
    if (timelineRef.value) {
      timelineRef.value.scrollTop = timelineRef.value.scrollHeight;
    }
  });
};

const focusInput = () => {
  nextTick(() => {
    if (inputRef.value) inputRef.value.focus();
  });
};

const formatTime = (isoString) => {
  try {
    const date = new Date(isoString);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return '';
  }
};

watch(() => chatStore.activeRoomId, () => {
  scrollToBottom();
  focusInput();
});

watch(() => currentMessages.value, () => {
  scrollToBottom();
}, { deep: true });

onMounted(() => {
  scrollToBottom();
  focusInput();
});
</script>

<style scoped>
.chat-window {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: var(--bg-primary);
}

.chat-header {
  height: 48px;
  padding: 0 16px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  align-items: center;
  gap: 8px;
}

.room-icon {
  font-size: 20px;
}

.room-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-header);
}

.messages-timeline {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.no-messages {
  text-align: center;
  color: var(--text-muted);
  font-size: 14px;
  margin-top: 40px;
}

.message-wrapper {
  display: flex;
  gap: 16px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.1s ease;
}

.message-wrapper:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: var(--accent-blurple);
  display: flex;
  justify-content: center;
  align-items: center;
  color: white;
  font-weight: 600;
  font-size: 16px;
  flex-shrink: 0;
}

.message-spacer {
  width: 40px;
  flex-shrink: 0;
}

.message-content-area {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.message-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.sender-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--text-header);
}

.sender-name.me {
  color: var(--text-link);
}

.send-time {
  font-size: 11px;
  color: var(--text-muted);
}

.message-text {
  font-size: 15px;
  line-height: 1.4;
  color: var(--text-normal);
  word-break: break-all;
  white-space: pre-wrap;
}

.message-wrapper.same-sender {
  margin-top: -4px;
  padding-top: 2px;
  padding-bottom: 2px;
}

.input-bar {
  padding: 16px;
  background-color: var(--bg-primary);
}

.input-form {
  display: flex;
  background-color: var(--bg-secondary);
  border-radius: 8px;
  padding: 4px 8px;
  align-items: center;
}

.chat-input {
  flex: 1;
  background: none;
  border: none;
  padding: 10px;
  outline: none;
  font-size: 15px;
  color: var(--text-normal);
}

.chat-input::placeholder {
  color: var(--text-muted);
}

.send-btn {
  background-color: var(--accent-blurple);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
  transition: var(--transition-smooth);
}

.send-btn:hover {
  background-color: var(--accent-blurple-hover);
}

.send-btn:disabled {
  background-color: var(--border-color);
  color: var(--text-muted);
  cursor: not-allowed;
}
</style>
