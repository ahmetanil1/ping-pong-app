package service

import (
	"context"

	"ping-service/internal/client"
)

type PingService interface {
	ExecutePingPong(ctx context.Context) error
}

type pingService struct {
	pongClient client.PongClient
}

func NewPingService(pongClient client.PongClient) PingService {
	return &pingService{pongClient: pongClient}
}

func (s *pingService) ExecutePingPong(ctx context.Context) error {
	// Business Logic: Pong client'ını çağırıp akışı kontrol eder
	return s.pongClient.SendPongRequest(ctx)
}