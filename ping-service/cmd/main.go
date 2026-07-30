package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ping-service/internal/client"
	"ping-service/internal/config"
	"ping-service/internal/handler"
	"ping-service/internal/service"
)

func main() {
	// 1. JSON Logger Ayarlanması
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Config Yüklenmesi
	cfg := config.Load()

	// 3. Dependency Injection
	pongClient := client.NewPongClient(cfg.PongURL)
	pingService := service.NewPingService(pongClient)
	pingHandler := handler.NewPingHandler(pingService)

	// 4. Router ve Server
	mux := http.NewServeMux()
	mux.HandleFunc("/", pingHandler.RenderUI)
	mux.HandleFunc("/api/v1/ping", pingHandler.Ping)
	mux.HandleFunc("/healthz/liveness", pingHandler.Liveness)
	mux.HandleFunc("/healthz/readiness", pingHandler.Readiness)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}

	// 5. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("ping-service baslatildi", "port", cfg.ServerPort, "pong_url", cfg.PongURL)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server durduruldu", "error", err)
			os.Exit(1)
		}
	}()

	<-stop
	slog.Info("ping-service kapatma sinyali aldi...")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		slog.Error("Graceful shutdown hatasi", "error", err)
	}
	slog.Info("ping-service guvenli bir sekilde durduruldu")
}
