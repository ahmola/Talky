<template>
  <div v-if="!authStore.isAuthenticated" class="auth-container">
    <div v-if="isRegisterMode" class="auth-card">
      <h2>계정 만들기</h2>
      <p>Talking에 가입하고 친구들과 소통해보세요</p>
      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label>사용자 아이디</label>
          <input v-model="regForm.username" type="text" class="form-input" required autocomplete="username" placeholder="아이디 입력" />
        </div>
        <div class="form-group">
          <label>닉네임</label>
          <input v-model="regForm.nickname" type="text" class="form-input" required placeholder="닉네임 입력" />
        </div>
        <div class="form-group">
          <label>비밀번호</label>
          <div class="input-wrapper">
            <input 
              v-model="regForm.password" 
              :type="showRegPassword ? 'text' : 'password'" 
              class="form-input" 
              required 
              autocomplete="new-password" 
              placeholder="비밀번호 입력"
            />
            <button type="button" class="pw-toggle-btn" @click="showRegPassword = !showRegPassword" :title="showRegPassword ? '비밀번호 숨기기' : '비밀번호 보기'">
              {{ showRegPassword ? '👁️' : '🔒' }}
            </button>
          </div>
        </div>
        <div class="form-group">
          <label>비밀번호 확인</label>
          <div class="input-wrapper">
            <input 
              v-model="regForm.passwordConfirm" 
              :type="showRegConfirm ? 'text' : 'password'" 
              class="form-input" 
              required 
              autocomplete="new-password" 
              placeholder="비밀번호 재입력"
            />
            <button type="button" class="pw-toggle-btn" @click="showRegConfirm = !showRegConfirm" :title="showRegConfirm ? '비밀번호 숨기기' : '비밀번호 보기'">
              {{ showRegConfirm ? '👁️' : '🔒' }}
            </button>
          </div>
          <p v-if="regForm.password && regForm.passwordConfirm && regForm.password !== regForm.passwordConfirm" class="error-text">
            비밀번호가 일치하지 않습니다.
          </p>
        </div>
        <button type="submit" class="auth-btn" :disabled="isSubmitting">
          {{ isSubmitting ? '가입 처리 중...' : '가입하기' }}
        </button>
      </form>
      <p class="auth-switch-text">
        이미 계정이 있으신가요?
        <span class="auth-link" @click="switchToLogin">로그인</span>
      </p>
    </div>
    <div v-else class="auth-card">
      <h2>돌아온 것을 환영합니다!</h2>
      <p>다시 만나서 반가워요!</p>
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>아이디</label>
          <input v-model="loginForm.username" type="text" class="form-input" required autocomplete="username" placeholder="아이디 입력" />
        </div>
        <div class="form-group">
          <label>비밀번호</label>
          <div class="input-wrapper">
            <input 
              v-model="loginForm.password" 
              :type="showLoginPassword ? 'text' : 'password'" 
              class="form-input" 
              required 
              autocomplete="current-password" 
              placeholder="비밀번호 입력"
            />
            <button type="button" class="pw-toggle-btn" @click="showLoginPassword = !showLoginPassword" :title="showLoginPassword ? '비밀번호 숨기기' : '비밀번호 보기'">
              {{ showLoginPassword ? '👁️' : '🔒' }}
            </button>
          </div>
        </div>
        <button type="submit" class="auth-btn" :disabled="isSubmitting">
          {{ isSubmitting ? '로그인 중...' : '로그인' }}
        </button>
      </form>
      <p class="auth-switch-text">
        계정이 필요하신가요?
        <span class="auth-link" @click="switchToRegister">가입하기</span>
      </p>
    </div>
  </div>
  <div v-else class="app-layout">
    <!-- 좌측 세로 탭바 (Discord 느낌 사이드바) -->
    <Sidebar :activeTab="activeTab" @changeTab="activeTab = $event" />
    
    <!-- 중간 목록 영역 (KakaoTalk 목록) -->
    <div class="list-section">
      <FriendList v-if="activeTab === 'friends'" />
      <ChatList v-else-if="activeTab === 'chats'" />
      <div v-else-if="activeTab === 'settings'" class="settings-view">
        <div class="settings-header">
          <h2>설정</h2>
        </div>
        <div class="settings-body">
          <div class="user-profile-large">
            <div class="avatar-large">{{ authStore.nickname ? authStore.nickname[0] : 'U' }}</div>
            <div class="profile-info">
              <h3>{{ authStore.nickname }}</h3>
              <p>@{{ authStore.username || 'user' }}</p>
            </div>
          </div>
          <button class="logout-btn" @click="authStore.logout()">로그아웃</button>
        </div>
      </div>
    </div>

    <!-- 우측 메인 채팅 영역 -->
    <div class="content-section">
      <ChatWindow v-if="chatStore.activeRoomId" />
      <div v-else class="empty-view">
        <div class="empty-logo">💬</div>
        <h2>친구와 대화를 시작해보세요!</h2>
        <p>친구 목록에서 1:1 대화를 시작하거나 그룹 채팅방을 개설할 수 있습니다.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue';
import { useAuthStore } from './store/auth';
import { useChatStore } from './store/chat';
import Sidebar from './components/Sidebar.vue';
import FriendList from './components/FriendList.vue';
import ChatList from './components/ChatList.vue';
import ChatWindow from './components/ChatWindow.vue';

const authStore = useAuthStore();
const chatStore = useChatStore();

const isRegisterMode = ref(false);
const activeTab = ref('friends'); // 'friends', 'chats', 'settings'
const isSubmitting = ref(false);

const showLoginPassword = ref(false);
const showRegPassword = ref(false);
const showRegConfirm = ref(false);

const loginForm = reactive({ username: '', password: '' });
const regForm = reactive({ username: '', password: '', passwordConfirm: '', nickname: '' });

const switchToLogin = () => {
  if (regForm.username) {
    loginForm.username = regForm.username.trim();
  }
  isRegisterMode.value = false;
};

const switchToRegister = () => {
  if (loginForm.username) {
    regForm.username = loginForm.username.trim();
  }
  isRegisterMode.value = true;
};

onMounted(() => {
  authStore.initialize();
});

const handleLogin = async () => {
  const username = loginForm.username.trim();
  if (!username) {
    alert('아이디를 입력해주세요.');
    return;
  }
  isSubmitting.value = true;
  try {
    await authStore.login(username, loginForm.password);
  } catch (err) {
    alert(err.message);
  } finally {
    isSubmitting.value = false;
  }
};

const handleRegister = async () => {
  const username = regForm.username.trim();
  const nickname = regForm.nickname.trim();
  const password = regForm.password;

  if (!username) {
    alert('아이디를 입력해주세요.');
    return;
  }
  if (!password) {
    alert('비밀번호를 입력해주세요.');
    return;
  }
  if (password !== regForm.passwordConfirm) {
    alert('비밀번호와 비밀번호 확인이 일치하지 않습니다.');
    return;
  }
  if (!nickname) {
    alert('닉네임을 입력해주세요.');
    return;
  }

  isSubmitting.value = true;
  try {
    await authStore.register(username, password, nickname);
    alert('가입이 완료되었습니다! 자동으로 로그인합니다.');
    await authStore.login(username, password);
  } catch (err) {
    alert(err.message);
  } finally {
    isSubmitting.value = false;
  }
};
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-primary);
}

.list-section {
  width: 320px;
  background-color: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.content-section {
  flex: 1;
  background-color: var(--bg-primary);
  display: flex;
  flex-direction: column;
}

.empty-view {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: var(--text-muted);
  text-align: center;
  padding: 40px;
}

.empty-logo {
  font-size: 72px;
  margin-bottom: 24px;
  animation: bounce 2s infinite ease-in-out;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.empty-view h2 {
  color: var(--text-header);
  font-size: 24px;
  margin-bottom: 8px;
}

.empty-view p {
  max-width: 380px;
  font-size: 15px;
}

/* 설정 뷰 */
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.settings-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  height: 48px;
  display: flex;
  align-items: center;
}

.settings-header h2 {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-header);
}

.settings-body {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.user-profile-large {
  display: flex;
  align-items: center;
  gap: 16px;
  background-color: var(--bg-tertiary);
  padding: 16px;
  border-radius: 8px;
}

.avatar-large {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background-color: var(--accent-blurple);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 24px;
  font-weight: 600;
  color: white;
}

.profile-info h3 {
  color: var(--text-header);
  font-size: 18px;
}

.profile-info p {
  color: var(--text-muted);
  font-size: 14px;
}

.logout-btn {
  padding: 10px;
  background-color: var(--accent-red);
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition-smooth);
}

.logout-btn:hover {
  background-color: #a62a2d;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-wrapper .form-input {
  padding-right: 42px;
}

.pw-toggle-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  padding: 4px;
  opacity: 0.7;
  transition: opacity 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.pw-toggle-btn:hover {
  opacity: 1;
}

.error-text {
  color: var(--accent-red);
  font-size: 12px;
  margin-top: 4px;
  text-align: left;
}
</style>
