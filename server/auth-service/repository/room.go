package repository

import (
	"context"
	"database/sql"
	"time"
)

type Room struct {
	ID        string    `json:"roomId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Message struct {
	ID             string    `json:"messageId"`
	RoomID         string    `json:"roomId"`
	SenderID       string    `json:"senderId"`
	SenderNickname string    `json:"senderNickname"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (d *Database) CreateRoom(ctx context.Context, name, roomType string, memberIDs []string) (string, error) {
	tx, err := d.Conn.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// 1. 방 생성
	var roomID string
	queryRoom := `INSERT INTO rooms (name, type) VALUES ($1, $2) RETURNING id`
	err = tx.QueryRowContext(ctx, queryRoom, name, roomType).Scan(&roomID)
	if err != nil {
		return "", err
	}

	// 2. 방 참여자 추가
	queryMember := `INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`
	for _, memberID := range memberIDs {
		_, err = tx.ExecContext(ctx, queryMember, roomID, memberID)
		if err != nil {
			return "", err
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return roomID, nil
}

func (d *Database) GetRoomsForUser(ctx context.Context, userID string) ([]Room, error) {
	query := `
		SELECT r.id, r.name, r.type, r.created_at, r.updated_at
		FROM rooms r
		JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1
		ORDER BY r.updated_at DESC
	`
	rows, err := d.Conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Room
	for rows.Next() {
		var r Room
		if err := rows.Scan(&r.ID, &r.Name, &r.Type, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	if list == nil {
		list = []Room{}
	}
	return list, nil
}

func (d *Database) GetRoomMembers(ctx context.Context, roomID string) ([]string, error) {
	query := `SELECT user_id FROM room_members WHERE room_id = $1`
	rows, err := d.Conn.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (d *Database) SaveMessage(ctx context.Context, roomID, senderID, content string) (*Message, error) {
	tx, err := d.Conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. 메시지 저장
	var msgID string
	var createdAt time.Time
	queryMsg := `INSERT INTO messages (room_id, sender_id, content) VALUES ($1, $2, $3) RETURNING id, created_at`
	err = tx.QueryRowContext(ctx, queryMsg, roomID, senderID, content).Scan(&msgID, &createdAt)
	if err != nil {
		return nil, err
	}

	// 2. 채팅방의 updated_at 갱신
	queryUpdateRoom := `UPDATE rooms SET updated_at = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, queryUpdateRoom, createdAt, roomID)
	if err != nil {
		return nil, err
	}

	// 3. 보낸 사람 닉네임 조회
	var nickname string
	queryNickname := `SELECT nickname FROM users WHERE id = $1`
	err = tx.QueryRowContext(ctx, queryNickname, senderID).Scan(&nickname)
	if err != nil {
		nickname = "Unknown"
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:             msgID,
		RoomID:         roomID,
		SenderID:       senderID,
		SenderNickname: nickname,
		Content:        content,
		CreatedAt:      createdAt,
	}, nil
}

func (d *Database) GetRoomMessages(ctx context.Context, roomID string, limit int, before time.Time) ([]Message, error) {
	var rows *sql.Rows
	var err error

	if before.IsZero() {
		query := `
			SELECT m.id, m.room_id, m.sender_id, u.nickname, m.content, m.created_at
			FROM messages m
			LEFT JOIN users u ON m.sender_id = u.id
			WHERE m.room_id = $1
			ORDER BY m.created_at DESC
			LIMIT $2
		`
		rows, err = d.Conn.QueryContext(ctx, query, roomID, limit)
	} else {
		query := `
			SELECT m.id, m.room_id, m.sender_id, u.nickname, m.content, m.created_at
			FROM messages m
			LEFT JOIN users u ON m.sender_id = u.id
			WHERE m.room_id = $1 AND m.created_at < $2
			ORDER BY m.created_at DESC
			LIMIT $3
		`
		rows, err = d.Conn.QueryContext(ctx, query, roomID, before, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Message
	for rows.Next() {
		var m Message
		var nick sql.NullString
		if err := rows.Scan(&m.ID, &m.RoomID, &m.SenderID, &nick, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		if nick.Valid {
			m.SenderNickname = nick.String
		} else {
			m.SenderNickname = "Unknown"
		}
		list = append(list, m)
	}
	if list == nil {
		list = []Message{}
	}

	// UI에서는 과거 -> 현재 순서로 타임라인이 그려져야 함
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}

	return list, nil
}
