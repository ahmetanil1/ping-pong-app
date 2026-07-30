package repository

import (
	"database/sql"
	"time"
)

type LogRepository interface {
	InsertLog(createdAt time.Time) error
	Ping() error
}

type postgresLogRepository struct {
	db *sql.DB
}

func NewPostgresLogRepository(db *sql.DB) LogRepository {
	return &postgresLogRepository{db: db}
}

func (r *postgresLogRepository) InsertLog(createdAt time.Time) error {
	_, err := r.db.Exec("INSERT INTO request_logs (created_at) VALUES ($1)", createdAt)
	return err
}

func (r *postgresLogRepository) Ping() error {
	return r.db.Ping()
}
