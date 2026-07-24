# 데이터베이스 스키마 및 인덱스 설계 명세

본 문서는 실시간 메신저 Talking에서 사용하는 서버용 PostgreSQL 데이터베이스와 클라이언트 캐시용 SQLite 데이터베이스의 스키마 및 쿼리 최적화를 위한 인덱스 설계 내용입니다.

---

## 1. 서버 데이터베이스 (PostgreSQL)

서버는 관계형 및 무결성을 보장하고, 인덱스를 활용해 빠른 친구 및 채팅방 조회, 메시지 조회를 보장하도록 설계합니다.

### 1.1. 테이블 상세 정의

#### 1.1.1. `users` (사용자 테이블)
| 컬럼명 | 타입 | 제약조건 | 설명 |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | 고유 사용자 식별자 |
| `username` | VARCHAR(50) | UNIQUE, NOT NULL | 로그인용 아이디 |
| `password_hash` | TEXT | NOT NULL | 단방향 해시된 비밀번호 |
| `nickname` | VARCHAR(50) | NOT NULL | 화면 표시용 이름 |
| `status_message` | VARCHAR(255) | DEFAULT '' | 프로필 상태 메시지 |
| `created_at` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 가입 일시 |

* **인덱스 최적화:**
  - `idx_users_username`: `username` 컬럼에 UNIQUE 인덱스를 설정하여 로그인 및 친구 추가 시 유저 검색 속도를 최소화합니다.

#### 1.1.2. `friends` (친구 관계 테이블)
사용자 A와 사용자 B의 관계를 양방향 또는 단방향 행으로 표현합니다. 관계는 상호 친구 상태를 뜻합니다.
| 컬럼명 | 타입 | 제약조건 | 설명 |
| :--- | :--- | :--- | :--- |
| `user_id` | UUID | FOREIGN KEY REFERENCES users(id) ON DELETE CASCADE | 기준 유저 |
| `friend_id` | UUID | FOREIGN KEY REFERENCES users(id) ON DELETE CASCADE | 대상 유저 (친구) |
| `created_at` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 친구 추가 일시 |

* **제약 조건:**
  - `PRIMARY KEY (user_id, friend_id)`: 중복된 친구 관계가 생성되는 것을 방지합니다.
* **인덱스 최적화:**
  - `idx_friends_friend_id`: 복합 기본키에 의해 `user_id` 기준 정렬 및 조회는 빠르나, 반대로 특정 유저를 누가 친구로 추가했는지 역방향 조회를 최적화하기 위해 `friend_id` 단독 인덱스를 생성합니다.

#### 1.1.3. `rooms` (채팅방 테이블)
| 컬럼명 | 타입 | 제약조건 | 설명 |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | 고유 채팅방 식별자 |
| `name` | VARCHAR(100) | NOT NULL | 채팅방 이름 (1:1인 경우 상대방 이름 등으로 대체 가능) |
| `type` | VARCHAR(20) | NOT NULL | `direct`(1:1) 또는 `group`(그룹) |
| `created_at` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 생성 일시 |
| `updated_at` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 마지막 메시지 전송 등 갱신 일시 |

#### 1.1.4. `room_members` (채팅방 참여자 테이블)
| 컬럼명 | 타입 | 제약조건 | 설명 |
| :--- | :--- | :--- | :--- |
| `room_id` | UUID | FOREIGN KEY REFERENCES rooms(id) ON DELETE CASCADE | 채팅방 ID |
| `user_id` | UUID | FOREIGN KEY REFERENCES users(id) ON DELETE CASCADE | 참여 사용자 ID |
| `joined_at` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 입장 일시 |

* **제약 조건:**
  - `PRIMARY KEY (room_id, user_id)`: 중복 참여를 방지합니다.
* **인덱스 최적화:**
  - `idx_room_members_user_id`: 특정 유저가 참여 중인 채팅방 목록을 빠르게 조회하기 위해 `user_id` 단독 인덱스를 설정합니다.

#### 1.1.5. `messages` (채팅 메시지 테이블)
| 컬럼명 | 타입 | 제약조건 | 설명 |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PRIMARY KEY, DEFAULT gen_random_uuid() | 고유 메시지 ID |
| `room_id` | UUID | FOREIGN KEY REFERENCES rooms(id) ON DELETE CASCADE | 소속 채팅방 ID |
| `sender_id` | UUID | FOREIGN KEY REFERENCES users(id) ON DELETE SET NULL | 발신 사용자 ID |
| `content` | TEXT | NOT NULL | 메시지 내용 |
| `created_at` | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | 전송 일시 |

* **인덱스 최적화:**
  - `idx_messages_room_created`: **`room_id`와 `created_at DESC` 복합 인덱스**를 생성합니다. 채팅방 입장 시 가장 최근 메시지 N개를 페이징 조회할 때 디스크 스캔량을 대폭 절감하고, 정렬 연산을 생략하여 성능을 극대화합니다.

---

## 2. 클라이언트 데이터베이스 (SQLite)

클라이언트는 브라우저 내부 WASM SQLite를 통해 오프라인 데이터 및 로컬 캐싱을 관리합니다. 서버 스키마의 일부분을 복사하며, 모바일/데스크톱 성능을 보장할 수 있도록 경량화된 구조를 가집니다.

### 2.1. 테이블 상세 정의

#### 2.1.1. `users` (친구 및 유저 캐시)
```sql
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    nickname TEXT NOT NULL,
    status_message TEXT DEFAULT '',
    online INTEGER DEFAULT 0 -- 0: Offline, 1: Online
);
```

#### 2.1.2. `friends` (로컬 친구 목록)
```sql
CREATE TABLE IF NOT EXISTS friends (
    user_id TEXT,
    friend_id TEXT,
    PRIMARY KEY (user_id, friend_id),
    FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### 2.1.3. `rooms` (로컬 채팅방 캐시)
```sql
CREATE TABLE IF NOT EXISTS rooms (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'direct' or 'group'
    updated_at TEXT NOT NULL -- ISO8601 String
);
```

#### 2.1.4. `messages` (로컬 메시지 히스토리 캐시)
```sql
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    room_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    sender_nickname TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TEXT NOT NULL, -- ISO8601 String
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);

-- 인덱스 설계: 특정 방의 메시지를 시간 역순으로 조회하는 쿼리 최적화
CREATE INDEX IF NOT EXISTS idx_local_messages_room_time ON messages (room_id, created_at DESC);
```
