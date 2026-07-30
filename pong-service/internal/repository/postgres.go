package repository

import (
    "database/sql"
    "log/slog"
    "time"

    _ "github.com/lib/pq"
)

type LogRepository interface {
    InsertLog(createdAt time.Time) error
    Ping() error
}

type postgresLogRepository struct {
    db *sql.DB
}

func NewPostgresLogRepository(db *sql.DB) LogRepository {
    db.SetMaxOpenConns(25)                 // Aynı anda açılabilecek maksimum bağlantı sayısı
    db.SetMaxIdleConns(25)                 // Havuzda boşta (idle) hazır tutulacak bağlantı sayısı
    db.SetConnMaxLifetime(5 * time.Minute) // Bir bağlantının maksimum ömrü
    db.SetConnMaxIdleTime(1 * time.Minute) // Boştaki bağlantının havuzda kalma süresi

    return &postgresLogRepository{db: db}
}

func (r *postgresLogRepository) InsertLog(createdAt time.Time) error {
    res, err := r.db.Exec("INSERT INTO request_logs (created_at) VALUES ($1)", createdAt)
    if err != nil {
        slog.Error("Postgres INSERT hatasi", "error", err)
        return err
    }

    rows, err := res.RowsAffected()
    if err != nil {
        slog.Error("RowsAffected alinamadi", "error", err)
    } else {
        slog.Info("Postgres INSERT basarili", "etkilenen_satir", rows)
    }

    return nil
}

func (r *postgresLogRepository) Ping() error {
    return r.db.Ping()
}