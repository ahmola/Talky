-- SQLite (클라이언트 브라우저 로컬 캐시용) 초기화 스키마 스크립트

-- 1. 로컬 유저 및 친구 상세 정보 캐시 테이블
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    nickname TEXT NOT NULL,
    status_message TEXT DEFAULT '',
    online INTEGER DEFAULT 0 -- 0: Offline, 1: Online
);

-- 2. 로컬 친구 매핑 테이블
CREATE TABLE IF NOT EXISTS friends (
    user_id TEXT,
    friend_id TEXT,
    PRIMARY KEY (user_id, friend_id),
    FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 3. 로컬 채팅방 테이블 캐시
CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'direct' or 'group'
    updated_at TEXT NOT NULL
);

-- 4. 로컬 메시지 히스토리 캐시 테이블
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    sender_nickname TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);

-- 특정 룸 내 메시지 시간역순 페이징 조회를 최적화하기 위한 로컬 복합 인덱스
CREATE INDEX IF NOT EXISTS idx_local_messages_room_time ON messages (room_id, created_at DESC);
