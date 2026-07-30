package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"pong-service/internal/config"
	"pong-service/internal/handler"
	"pong-service/internal/repository"
	"pong-service/internal/service"
)

func main() {
	// log/slog
	// JSON formatında loglama için slog kütüphanesini kullanıyoruz
	// Logları plain text yerine key-value formatında tutar. İleride log yönetim sistemleri (ELK, Prometheus, Grafana) ile entegre edilebilir

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Config Yüklenmesi
	cfg := config.Load()

	// 3. Altyapı Bağlantıları
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("PostgreSQL baglanti hatasi", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
	})
	defer rdb.Close()

	createTable(db)

	// 4. Dependency Injection
	logRepo := repository.NewPostgresLogRepository(db)
	cacheRepo := repository.NewRedisCacheRepository(rdb)

	pongService := service.NewPongService(logRepo, cacheRepo)
	pongHandler := handler.NewPongHandler(pongService)

	// 5. Router ve Server
	mux := http.NewServeMux()
	mux.HandleFunc("/internal/process", pongHandler.Process)
	mux.HandleFunc("/healthz/liveness", pongHandler.Liveness)
	mux.HandleFunc("/healthz/readiness", pongHandler.Readiness)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	// 6. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("pong-service baslatildi", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server durduruldu", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("pong-service kapatma sinyali aldi...")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		slog.Error("Graceful shutdown hatasi", "error", err)
	}
	slog.Info("pong-service guvenli bir sekilde durduruldu")
}

func createTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS request_logs (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP NOT NULL
	);`
	if _, err := db.Exec(query); err != nil {
		slog.Error("Veritabanı tablosu olusturulamadi", "error", err)
		os.Exit(1)
	}
}
