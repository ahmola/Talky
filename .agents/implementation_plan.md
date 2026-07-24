# 채팅 서비스 (Talking) 구현 계획서 (수정본)

본 계획서는 사용자의 피드백을 반영하여 디렉터리 구조를 최적화하고, 웹서버를 Caddy로 교체하며, 데이터베이스 스키마/인덱스 문서화 및 SQL 복구/초기화 스크립트 추가 계획을 보완한 최종 구현 계획서입니다.

## User Review Required

> [!IMPORTANT]
> **1. 디렉터리 구조 최적화**
> - 불필요한 `chat` 하위 디렉터리 중첩을 제거하고, `server`, `client`, `docs`, `k6` 디렉터리를 프로젝트 루트(`talking/`)에 바로 위치시킵니다.
> 
> **2. 웹서버 Caddy 도입**
> - 프론트엔드 정적 파일 배포 및 백엔드 서비스(REST API, WebSocket)에 대한 리버스 프록싱을 담당할 웹서버로 **Caddy**를 사용합니다.
> 
> **3. SQL 디렉터리 및 스키마/인덱스 문서화 추가**
> - 장애나 예외 상황에 대비하여 데이터베이스를 재구축하거나 수동 쿼리를 실행할 수 있도록 `sql` 디렉터리에 PostgreSQL 및 SQLite용 SQL 스크립트를 관리합니다.
> - `docs/schema.md`를 추가하여 테이블 설계와 인덱스 최적화 전략을 상세히 기록합니다.

## Open Questions

- 피드백에 따라 모든 구조를 확정하였으며, 이 계획이 승인되는 대로 프로토콜 문서화(`docs/` 작성) 단계부터 본격적인 개발에 착수합니다.

---

## Proposed Changes

수정된 전체 프로젝트 구조는 다음과 같습니다:

```
c:\workspace\talking/
├── docker-compose.yml          # 전체 서비스 통합 실행 (Caddy, Postgres, auth, chat)
├── Caddyfile                   # Caddy 리버스 프록시 및 static 파일 서빙 설정
├── docs/                       # 명세 및 문서화 디렉터리
│   ├── openapi.yaml            # REST API 명세
│   ├── websocket.md            # WebSocket 프로토콜 명세
│   ├── grpc.md                 # gRPC 프로토콜 및 메서드 명세
│   └── schema.md               # [NEW] DB 스키마 및 인덱스 최적화 설계 문서
├── server/                     # Go 백엔드
│   ├── proto/                  # gRPC Protobuf 정의 (.proto 및 생성 파일)
│   ├── auth-service/           # 사용자/친구 관리 및 gRPC 서버
│   └── chat-service/           # WebSocket 중계 및 gRPC 클라이언트
├── client/                     # Vue.js 프론트엔드 (WASM SQLite 연동)
│   └── src/
│       ├── db/                 # WASM SQLite 연동 모듈
│       ├── store/              # Pinia 상태 관리
│       └── components/         # KakaoTalk 구조 + Discord 테마 UI
├── sql/                        # [NEW] DB 에러 대비 및 초기화 SQL 스크립트
│   ├── postgres/               # PostgreSQL DDL, DML, Index 생성 SQL
│   └── sqlite/                 # SQLite DDL, Index 생성 SQL
└── k6/                         # k6 부하 테스트 시나리오 (TS/JS)
```

---

### 1. Protocol & DB Schema Specifications (`docs/`)

#### [NEW] [openapi.yaml](file:///c:/workspace/talking/docs/openapi.yaml)
- 회원가입, 로그인, 친구 목록 CRUD, 채팅방 정보 조회 등 REST API 명세를 기술합니다.

#### [NEW] [websocket.md](file:///c:/workspace/talking/docs/websocket.md)
- WebSocket 연결 흐름, 메시지 송수신 JSON 포맷에 대한 명세를 기술합니다.

#### [NEW] [grpc.md](file:///c:/workspace/talking/docs/grpc.md)
- `chat-service`와 `auth-service` 간의 유저 검증, 친구 관계 확인 등을 위한 gRPC 인터페이스 명세를 작성합니다.

#### [NEW] [schema.md](file:///c:/workspace/talking/docs/schema.md)
- PostgreSQL 및 SQLite의 테이블 구조, 관계도, 성능 최적화를 위한 인덱스 설계 내용을 상세히 기록합니다.

---

### 2. SQL Scripts (`sql/`)

#### [NEW] [postgres/](file:///c:/workspace/talking/sql/postgres)
- PostgreSQL의 초기 테이블 생성 쿼리, 인덱스 생성 쿼리 및 기본 더미 데이터 삽입 쿼리 등을 모아둡니다. (`schema.sql`, `indices.sql` 등)

#### [NEW] [sqlite/](file:///c:/workspace/talking/sql/sqlite)
- 클라이언트 웹 브라우저 내에서 SQLite를 초기화할 때 실행할 DDL 및 인덱스 정의 SQL 쿼리를 관리합니다.

---

### 3. Backend & Frontend Architecture

#### [NEW] [server/auth-service](file:///c:/workspace/talking/server/auth-service)
- **인터페이스 추상화**: 로그인 및 세션 관리 기능을 추상화하여, 추후 JWT/OAuth2 프로토콜 적용이 용이하도록 구성합니다.
- Go, PostgreSQL, gRPC 서버 구현.

#### [NEW] [server/chat-service](file:///c:/workspace/talking/server/chat-service)
- Gorilla WebSocket 기반의 실시간 중계 브로커.
- WebSocket 연결 후 클라이언트 검증을 위해 `auth-service`와 gRPC로 실시간 연동합니다.

#### [NEW] [client](file:///c:/workspace/talking/client)
- Vue.js 3, Pinia 상태 관리.
- `@vlcn.io/wa-sqlite`를 활용하여 브라우저 로컬 DB(SQLite) 구축 및 API 캐싱 레이어 구현.
- UI: Discord 다크 감성 테마 + KakaoTalk 내비게이션 및 대화방 목록 구조.

#### [NEW] [Caddyfile](file:///c:/workspace/talking/Caddyfile)
- Vue.js 빌드 산출물을 정적 서빙하고, `/api` 경로는 `auth-service`로, `/ws` 경로는 `chat-service`로 라우팅하는 Caddy 리버스 프록시 설정을 작성합니다.

---

## Verification Plan

### Automated Tests
1. **백엔드 단위 테스트**: `go test ./...` 실행을 통해 REST API 핸들러 및 gRPC 통신 동작 검증.
2. **부하 테스트**: k6를 실행하여 WebSocket 중계 서버의 부하 한계 테스트.
   ```bash
   k6 run k6/dist/load_test.js
   ```

### Manual Verification
1. `docker-compose up --build` 실행 후 Caddy 포트(예: 80)를 통해 접속하여 테스트 진행.
2. 브라우저 개발자 도구의 Application -> IndexedDB 탭에서 SQLite DB의 캐시 저장 상태 및 쿼리 동작을 검증.
