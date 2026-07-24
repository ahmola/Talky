package repository

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            string
	Username      string
	PasswordHash  string
	Nickname      string
	StatusMessage string
}

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func (d *Database) CreateUser(ctx context.Context, username, password, nickname string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	var userID string
	query := `INSERT INTO users (username, password_hash, nickname) VALUES ($1, $2, $3) RETURNING id`
	err = d.Conn.QueryRowContext(ctx, query, username, string(hashed), nickname).Scan(&userID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (d *Database) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, password_hash, nickname, status_message FROM users WHERE username = $1`
	var u User
	err := d.Conn.QueryRowContext(ctx, query, username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Nickname, &u.StatusMessage)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *Database) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `SELECT id, username, nickname, status_message FROM users WHERE id = $1`
	var u User
	err := d.Conn.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Username, &u.Nickname, &u.StatusMessage)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return &u, nil
}

func (d *Database) CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
