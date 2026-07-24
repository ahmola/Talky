# WebSocket 프로토콜 명세 (WebSocket Protocol Specification)

실시간 메시지 전송 및 이벤트를 처리하기 위한 WebSocket 프로토콜 규격입니다.

## 1. 연결 엔드포인트
- **URL**: `ws://<host>/ws`
- **인증**: WebSocket 핸드셰이크 시 Query Parameter로 JWT 토큰을 전달하거나, 연결 후 첫 번째 메시지로 인증 요청을 보냅니다. 본 구현에서는 간편함과 보안을 위해 Query Parameter 방식을 기본으로 채택합니다.
  - 예시: `ws://localhost/ws?token=<JWT_TOKEN>`

---

## 2. 메시지 기본 포맷
모든 WebSocket 프레임은 JSON 형식의 텍스트 메시지입니다. 모든 메시지는 공통적으로 `event` 필드를 가져야 합니다.

```json
{
  "event": "이벤트명",
  "data": { ... }
}
```

---

## 3. 클라이언트 -> 서버 이벤트 (Client-to-Server Events)

### 3.1. 메시지 전송 (`send_message`)
채팅방에 텍스트 메시지를 보냅니다.

* **요청 바디:**
```json
{
  "event": "send_message",
  "data": {
    "roomId": "room-uuid-1234",
    "content": "안녕하세요! 반갑습니다."
  }
}
```

---

## 4. 서버 -> 클라이언트 이벤트 (Server-to-Client Events)

### 4.1. 새로운 메시지 수신 (`new_message`)
채팅방의 다른 참여자가 보낸 메시지를 실시간으로 수신합니다.

* **이벤트 포맷:**
```json
{
  "event": "new_message",
  "data": {
    "messageId": "msg-uuid-5678",
    "roomId": "room-uuid-1234",
    "senderId": "user-uuid-9999",
    "senderNickname": "홍길동",
    "content": "안녕하세요! 반갑습니다.",
    "createdAt": "2026-07-24T14:40:00Z"
  }
}
```

### 4.2. 친구 온라인 상태 변경 (`status_change`)
친구 목록에 있는 유저의 접속 상태가 변경되었을 때 브로드캐스트됩니다.

* **이벤트 포맷:**
```json
{
  "event": "status_change",
  "data": {
    "userId": "user-uuid-9999",
    "online": true
  }
}
```

### 4.3. 에러 발생 (`error`)
클라이언트의 요청 처리 중 오류가 발생한 경우 전달됩니다.

* **이벤트 포맷:**
```json
{
  "event": "error",
  "data": {
    "code": "UNAUTHORIZED_ROOM",
    "message": "해당 채팅방에 접근할 권한이 없습니다."
  }
}
```
