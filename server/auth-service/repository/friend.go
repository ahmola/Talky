package repository

import (
	"context"
)

type FriendInfo struct {
	UserID        string `json:"userId"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	StatusMessage string `json:"statusMessage"`
}

func (d *Database) GetFriends(ctx context.Context, userID string) ([]FriendInfo, error) {
	query := `
		SELECT u.id, u.username, u.nickname, u.status_message
		FROM friends f
		JOIN users u ON f.friend_id = u.id
		WHERE f.user_id = $1
	`
	rows, err := d.Conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []FriendInfo
	for rows.Next() {
		var f FriendInfo
		if err := rows.Scan(&f.UserID, &f.Username, &f.Nickname, &f.StatusMessage); err != nil {
			return nil, err
		}
		list = append(list, f)
	}
	// Return empty array instead of nil for JSON consistency
	if list == nil {
		list = []FriendInfo{}
	}
	return list, nil
}

func (d *Database) AddFriend(ctx context.Context, userID, friendID string) error {
	tx, err := d.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// A -> B 추가
	query1 := `INSERT INTO friends (user_id, friend_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	if _, err := tx.ExecContext(ctx, query1, userID, friendID); err != nil {
		return err
	}

	// B -> A 추가
	if _, err := tx.ExecContext(ctx, query1, friendID, userID); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) DeleteFriend(ctx context.Context, userID, friendID string) error {
	tx, err := d.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM friends WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`
	if _, err := tx.ExecContext(ctx, query, userID, friendID); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM friends WHERE user_id = $1 AND friend_id = $2)`
	var exists bool
	err := d.Conn.QueryRowContext(ctx, query, userID, friendID).Scan(&exists)
	return exists, err
}
