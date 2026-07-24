<template>
  <div class="sidebar">
    <!-- 로고 영역 -->
    <div class="logo-area">
      <div class="logo-icon">T</div>
    </div>
    
    <div class="divider"></div>

    <!-- 네비게이션 탭 메뉴 -->
    <div class="nav-items">
      <button 
        class="nav-btn" 
        :class="{ active: activeTab === 'friends' }"
        @click="$emit('changeTab', 'friends')"
        title="친구"
      >
        <span class="btn-icon">👥</span>
      </button>
      
      <button 
        class="nav-btn" 
        :class="{ active: activeTab === 'chats' }"
        @click="$emit('changeTab', 'chats')"
        title="채팅"
      >
        <span class="btn-icon">💬</span>
      </button>
    </div>

    <!-- 설정 영역 -->
    <div class="bottom-area">
      <button 
        class="nav-btn" 
        :class="{ active: activeTab === 'settings' }"
        @click="$emit('changeTab', 'settings')"
        title="설정"
      >
        <span class="btn-icon">⚙️</span>
      </button>
    </div>
  </div>
</template>

<script setup>
defineProps({
  activeTab: {
    type: String,
    required: true
  }
});
defineEmits(['changeTab']);
</script>

<style scoped>
.sidebar {
  width: 72px;
  background-color: var(--bg-tertiary);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
  gap: 8px;
}

.logo-area {
  display: flex;
  justify-content: center;
  align-items: center;
}

.logo-icon {
  width: 48px;
  height: 48px;
  background-color: var(--accent-blurple);
  border-radius: 50%;
  color: white;
  font-size: 22px;
  font-weight: 700;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  transition: var(--transition-smooth);
}

.logo-icon:hover {
  border-radius: 16px;
  background-color: var(--accent-blurple-hover);
}

.divider {
  width: 32px;
  height: 2px;
  background-color: var(--border-color);
  margin: 4px 0;
}

.nav-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.nav-btn {
  width: 48px;
  height: 48px;
  background-color: var(--bg-primary);
  border: none;
  border-radius: 50%;
  color: var(--text-normal);
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  position: relative;
  transition: var(--transition-smooth);
}

.nav-btn:hover {
  border-radius: 16px;
  background-color: var(--accent-blurple);
  color: white;
}

.nav-btn.active {
  border-radius: 16px;
  background-color: var(--accent-blurple);
  color: white;
}

/* Discord 스타일의 사이드바 액티브 인디케이터 */
.nav-btn::before {
  content: '';
  position: absolute;
  left: -12px;
  width: 4px;
  height: 8px;
  background-color: white;
  border-radius: 0 4px 4px 0;
  transition: var(--transition-smooth);
  opacity: 0;
}

.nav-btn:hover::before {
  opacity: 1;
  height: 20px;
}

.nav-btn.active::before {
  opacity: 1;
  height: 40px;
}

.btn-icon {
  font-size: 20px;
}

.bottom-area {
  margin-top: auto;
}
</style>
