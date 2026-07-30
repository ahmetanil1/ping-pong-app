package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"

	"ping-service/internal/dto"
	"ping-service/internal/service"
)

type PingHandler struct {
	service service.PingService
}

func NewPingHandler(service service.PingService) *PingHandler {
	return &PingHandler{service: service}
}

// UI Render eden Handler fonksiyonu
func (h *PingHandler) RenderUI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/index.html")
	if err != nil {
		slog.Error("Template yuklenemedi", "error", err)
		http.Error(w, "UI Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func (h *PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(dto.NewErrorResponse("Method not allowed"))
		return
	}

	if err := h.service.ExecutePingPong(r.Context()); err != nil {
		slog.Error("Ping-Pong akisi basarisiz", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.NewErrorResponse("Ping-Pong execution failed: " + err.Error()))
		return
	}

	slog.Info("Ping-Pong akisi basariyla tamamlandi")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.NewSuccessResponse("ping-pong completed successfully", nil))
}

func (h *PingHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *PingHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("READY"))
}
