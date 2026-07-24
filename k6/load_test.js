import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';

// k6 부하 테스트 옵션 설정 (가상 유저(VU) 20명을 유도하는 시나리오)
export const options = {
  stages: [
    { duration: '10s', target: 20 }, // 10초 동안 VU 20명으로 증가
    { duration: '20s', target: 20 }, // 20초 동안 VU 20명 유지
    { duration: '10s', target: 0 },  // 10초 동안 VU 0명으로 감소
  ],
};

const BASE_URL = 'http://gateway'; // Caddy Gateway 컨테이너/호스트 주소
const WS_URL = 'ws://gateway/ws';

export default function () {
  const rand = Math.floor(Math.random() * 1000000);
  const username = `k6user_${rand}`;
  const password = 'password123';
  const nickname = `k6tester_${rand}`;

  // 1. 회원가입 API 테스트
  let regRes = http.post(`${BASE_URL}/api/auth/register`, JSON.stringify({
    username: username,
    password: password,
    nickname: nickname,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(regRes, {
    'signup status is 201': (r) => r.status === 201,
  });

  // 2. 로그인 API 테스트
  let loginRes = http.post(`${BASE_URL}/api/auth/login`, JSON.stringify({
    username: username,
    password: password,
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  if (!check(loginRes, { 'login status is 200': (r) => r.status === 200 })) {
    return;
  }

  const token = loginRes.json().token;
  const authHeaders = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  // 3. REST API 친구 목록 및 대화방 목록 조회 테스트
  let friendsRes = http.get(`${BASE_URL}/api/friends`, { headers: authHeaders });
  check(friendsRes, { 'get friends status is 200': (r) => r.status === 200 });

  let roomsRes = http.get(`${BASE_URL}/api/rooms`, { headers: authHeaders });
  check(roomsRes, { 'get rooms status is 200': (r) => r.status === 200 });

  sleep(1);

  // 4. WebSocket 실시간 연결 및 메시징 부하 시뮬레이션
  const url = `${WS_URL}?token=${token}`;
  
  let wsRes = ws.connect(url, {}, function (socket) {
    socket.on('open', function () {
      check(socket, { 'websocket handshake established': (s) => true });

      // 연결 수립 1초 뒤 임의의 메시지 전송
      socket.setTimeout(function () {
        // 더미 룸 UUID 전달
        const mockRoomId = '00000000-0000-0000-0000-000000000000';
        socket.send(JSON.stringify({
          event: 'send_message',
          data: {
            roomId: mockRoomId,
            content: 'Hello from k6 load test client!'
          }
        }));
      }, 1000);

      // 연결 5초 뒤 자동 종료
      socket.setTimeout(function () {
        socket.close();
      }, 5000);
    });

    socket.on('message', function (data) {
      check(socket, { 'received message frame': (s) => data !== null });
    });

    socket.on('close', function () {
      // Close callback
    });

    socket.on('error', function (err) {
      // Error callback
    });
  });

  check(wsRes, {
    'websocket handshake status is 101': (r) => r.status === 101,
  });

  sleep(2);
}
