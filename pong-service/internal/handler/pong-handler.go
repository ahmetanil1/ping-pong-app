package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"pong-service/internal/dto"
	"pong-service/internal/service"
)

type PongHandler struct {
	service service.PongService
}

func NewPongHandler(service service.PongService) *PongHandler {
	return &PongHandler{service: service}
}

func (h *PongHandler) Process(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(dto.NewErrorResponse("Method not allowed"))
		return
	}

	if err := h.service.ProcessPong(r.Context()); err != nil {
		slog.Error("Pong islemi basarisiz oldu", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.NewErrorResponse("Failed to process pong: " + err.Error()))
		return
	}

	slog.Info("Pong istegi basariyla islendi")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.NewSuccessResponse("pong", nil))
}

func (h *PongHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *PongHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	if !h.service.CheckHealth(r.Context()) {
		slog.Warn("Readiness kontrolu basarisiz: DB veya Redis erişilemez durumda")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not Ready"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}
