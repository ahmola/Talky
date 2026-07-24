import sqlite3InitModule from '@sqlite.org/sqlite-wasm';

let db = null;

// SQLite WASM을 비동기로 로드하고 로컬 스키마를 초기화합니다.
export async function initDB() {
  if (db) return db;
  try {
    const sqlite3 = await sqlite3InitModule({
      print: console.log,
      printErr: console.error,
    });

    // 브라우저가 OPFS를 지원하면 영속성 디렉터리에 저장, 없으면 메모리/브라우저 휘발성 저장소 사용
    if ('opfs' in sqlite3) {
      db = new sqlite3.oo1.OpfsDb('/talking.db', 'c');
      console.log('SQLite initialized: OPFS mode');
    } else {
      db = new sqlite3.oo1.DB('/talking.db', 'c');
      console.warn('OPFS not supported. SQLite initialized in fallback memory mode.');
    }

    // DDL 실행
    db.exec(`
      CREATE TABLE IF NOT EXISTS users (
          id TEXT PRIMARY KEY,
          username TEXT NOT NULL,
          nickname TEXT NOT NULL,
          status_message TEXT DEFAULT '',
          online INTEGER DEFAULT 0
      );
      CREATE TABLE IF NOT EXISTS friends (
          user_id TEXT,
          friend_id TEXT,
          PRIMARY KEY (user_id, friend_id),
          FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE
      );
      CREATE TABLE IF NOT EXISTS rooms (
          id TEXT PRIMARY KEY,
          name TEXT NOT NULL,
          type TEXT NOT NULL,
          updated_at TEXT NOT NULL
      );
      CREATE TABLE IF NOT EXISTS messages (
          id TEXT PRIMARY KEY,
          room_id TEXT NOT NULL,
          sender_id TEXT NOT NULL,
          sender_nickname TEXT NOT NULL,
          content TEXT NOT NULL,
          created_at TEXT NOT NULL,
          FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
      );
      CREATE INDEX IF NOT EXISTS idx_local_messages_room_time ON messages (room_id, created_at DESC);
    `);

    return db;
  } catch (err) {
    console.error('Failed to init SQLite WASM:', err);
    throw err;
  }
}

export function getDB() {
  if (!db) {
    throw new Error('Database is not initialized. Call initDB() first.');
  }
  return db;
}

// ----------------------------------------------------
// 헬퍼 CRUD 쿼리 함수 정의
// ----------------------------------------------------

export function saveLocalUser(user) {
  const sqliteDB = getDB();
  sqliteDB.exec({
    sql: `INSERT INTO users (id, username, nickname, status_message, online) 
          VALUES ($id, $username, $nickname, $status_message, $online)
          ON CONFLICT(id) DO UPDATE SET 
            nickname=excluded.nickname,
            status_message=excluded.status_message,
            online=excluded.online`,
    bind: {
      $id: user.userId || user.id,
      $username: user.username,
      $nickname: user.nickname,
      $status_message: user.statusMessage || '',
      $online: user.online ? 1 : 0
    }
  });
}

export function getLocalFriends(currentUserID) {
  const sqliteDB = getDB();
  const result = [];
  sqliteDB.exec({
    sql: `SELECT u.id, u.username, u.nickname, u.status_message, u.online 
          FROM friends f
          JOIN users u ON f.friend_id = u.id
          WHERE f.user_id = $userID`,
    bind: { $userID: currentUserID },
    rowMode: 'object',
    callback: (row) => {
      result.push({
        userId: row.id,
        username: row.username,
        nickname: row.nickname,
        statusMessage: row.status_message,
        online: row.online === 1
      });
    }
  });
  return result;
}

export function saveLocalFriend(currentUserID, friend) {
  saveLocalUser(friend);
  const sqliteDB = getDB();
  sqliteDB.exec({
    sql: `INSERT OR IGNORE INTO friends (user_id, friend_id) VALUES ($userID, $friendID)`,
    bind: {
      $userID: currentUserID,
      $friendID: friend.userId || friend.id
    }
  });
}

export function clearFriends(currentUserID) {
  const sqliteDB = getDB();
  sqliteDB.exec({
    sql: `DELETE FROM friends WHERE user_id = $userID`,
    bind: { $userID: currentUserID }
  });
}

export function getLocalRooms() {
  const sqliteDB = getDB();
  const result = [];
  sqliteDB.exec({
    sql: `SELECT id, name, type, updated_at FROM rooms ORDER BY updated_at DESC`,
    rowMode: 'object',
    callback: (row) => {
      result.push({
        roomId: row.id,
        name: row.name,
        type: row.type,
        updatedAt: row.updated_at
      });
    }
  });
  return result;
}

export function saveLocalRoom(room) {
  const sqliteDB = getDB();
  sqliteDB.exec({
    sql: `INSERT INTO rooms (id, name, type, updated_at) 
          VALUES ($id, $name, $type, $updated_at)
          ON CONFLICT(id) DO UPDATE SET 
            name=excluded.name, 
            updated_at=excluded.updated_at`,
    bind: {
      $id: room.roomId || room.id,
      $name: room.name,
      $type: room.type,
      $updated_at: room.updatedAt || new Date().toISOString()
    }
  });
}

export function getLocalMessages(roomID, limit = 50) {
  const sqliteDB = getDB();
  const result = [];
  sqliteDB.exec({
    sql: `SELECT id, room_id, sender_id, sender_nickname, content, created_at 
          FROM messages 
          WHERE room_id = $roomId 
          ORDER BY created_at ASC 
          LIMIT $limit`,
    bind: {
      $roomId: roomID,
      $limit: limit
    },
    rowMode: 'object',
    callback: (row) => {
      result.push({
        messageId: row.id,
        roomId: row.room_id,
        senderId: row.sender_id,
        senderNickname: row.sender_nickname,
        content: row.content,
        createdAt: row.created_at
      });
    }
  });
  return result;
}

export function saveLocalMessage(msg) {
  const sqliteDB = getDB();
  sqliteDB.exec({
    sql: `INSERT INTO messages (id, room_id, sender_id, sender_nickname, content, created_at) 
          VALUES ($id, $roomId, $senderId, $senderNickname, $content, $createdAt)
          ON CONFLICT(id) DO NOTHING`,
    bind: {
      $id: msg.messageId || msg.id,
      $roomId: msg.roomId,
      $senderId: msg.senderId,
      $senderNickname: msg.senderNickname,
      $content: msg.content,
      $createdAt: msg.createdAt
    }
  });

  // 메시지 저장 시 로컬 방의 updated_at도 같이 업데이트
  sqliteDB.exec({
    sql: `UPDATE rooms SET updated_at = $updatedAt WHERE id = $roomId`,
    bind: {
      $roomId: msg.roomId,
      $updatedAt: msg.createdAt
    }
  });
}
