# gRPC 프로토콜 및 서비스 명세 (gRPC Protocol Specification)

`chat-service`(WebSocket 중계)와 `auth-service`(인증 및 유저 관리) 간의 신뢰성 있고 빠른 통신을 위한 내부 gRPC 프로토콜 정의입니다.

## 1. Protobuf 파일 위치 및 빌드
- **정의 파일**: `server/proto/auth.proto`
- **Go 빌드 결과물**: `server/proto/auth_grpc.pb.go`, `server/proto/auth.pb.go`

---

## 2. gRPC 서비스 인터페이스

### 2.1. `AuthService`
인증 서비스(`auth-service`)가 제공하며, 채팅 서비스(`chat-service`)가 클라이언트로서 호출합니다.

#### 서비스 정의

```protobuf
syntax = "proto3";

package auth;

option go_package = "talking/server/proto/auth";

service AuthService {
  // 클라이언트의 JWT 토큰을 검증하고 사용자 정보를 가져옵니다.
  rpc VerifyToken (VerifyTokenRequest) returns (VerifyTokenResponse);

  // 채팅방에 속해 있는 멤버들의 ID 목록을 가져옵니다. (메시지 브로드캐스팅용)
  rpc GetRoomMembers (GetRoomMembersRequest) returns (GetRoomMembersResponse);

  // 새로운 메시지를 데이터베이스에 저장합니다.
  rpc SaveMessage (SaveMessageRequest) returns (SaveMessageResponse);

  // 두 사용자가 친구 상태인지 검증합니다.
  rpc CheckFriendship (CheckFriendshipRequest) returns (CheckFriendshipResponse);
}
```

---

## 3. 메시지 구조체 명세

### 3.1. `VerifyToken`
- **Request**:
  - `token` (string): JWT 인증 토큰
- **Response**:
  - `userId` (string): 유저 UUID
  - `username` (string): 유저 아이디
  - `nickname` (string): 유저 닉네임

### 3.2. `GetRoomMembers`
- **Request**:
  - `roomId` (string): 채팅방 UUID
- **Response**:
  - `userIds` (repeated string): 해당 채팅방에 참여 중인 유저 UUID 목록

### 3.3. `SaveMessage`
- **Request**:
  - `roomId` (string): 채팅방 UUID
  - `senderId` (string): 송신자 UUID
  - `content` (string): 메시지 내용
- **Response**:
  - `messageId` (string): 생성된 메시지 UUID
  - `createdAt` (string): 메시지 생성 시간 (RFC3339 string)
  - `senderNickname` (string): 송신자 닉네임 (브로드캐스트용 편의 제공)

### 3.4. `CheckFriendship`
- **Request**:
  - `userId` (string): 기준 유저 UUID
  - `friendId` (string): 대상 유저 UUID
- **Response**:
  - `isFriend` (bool): 친구 여부
