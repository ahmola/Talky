-- PostgreSQL 초기화 스키마 및 인덱스 생성 스크립트

-- 1. 유저 테이블
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    status_message VARCHAR(255) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 유저네임 단방향 인덱스 (로그인 및 검색 최적화)
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);

-- 2. 친구 관계 테이블
CREATE TABLE IF NOT EXISTS friends (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, friend_id)
);

-- 역방향 친구 조회를 위한 인덱스
CREATE INDEX IF NOT EXISTS idx_friends_friend_id ON friends (friend_id);

-- 3. 채팅방 테이블
CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'direct', -- 'direct' or 'group'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. 채팅방 참여자 매핑 테이블
CREATE TABLE IF NOT EXISTS room_members (
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (room_id, user_id)
);

-- 유저가 속한 룸 목록 조회를 위한 인덱스
CREATE INDEX IF NOT EXISTS idx_room_members_user_id ON room_members (user_id);

-- 5. 메시지 테이블
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 특정 룸 내 최근 메시지 조회 성능 극대화를 위한 복합 인덱스 (Time-series paging)
CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages (room_id, created_at DESC);
