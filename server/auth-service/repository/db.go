package repository

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Database struct {
	Conn *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return &Database{Conn: db}, nil
}

func (d *Database) Close() error {
	return d.Conn.Close()
}

// InitSchema는 schema.sql 내용을 실행하여 DB 테이블 및 인덱스를 초기화합니다.
func (d *Database) InitSchema(schemaPath string) error {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file (%s): %w", schemaPath, err)
	}

	_, err = d.Conn.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute schema SQL: %w", err)
	}

	return nil
}
