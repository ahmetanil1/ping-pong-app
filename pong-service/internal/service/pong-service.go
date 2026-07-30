package service

import (
	"context"
	"time"

	"pong-service/internal/repository"
)

type PongService interface {
	ProcessPong(ctx context.Context) error
	CheckHealth(ctx context.Context) bool
}

type pongService struct {
	logRepo   repository.LogRepository
	cacheRepo repository.CacheRepository
}

func NewPongService(logRepo repository.LogRepository, cacheRepo repository.CacheRepository) PongService {
	return &pongService{
		logRepo:   logRepo,
		cacheRepo: cacheRepo,
	}
}

func (s *pongService) ProcessPong(ctx context.Context) error {
	now := time.Now()

	// 1. İş Mantığı: Log kaydını DB'ye at
	if err := s.logRepo.InsertLog(now); err != nil {
		return err
	}

	// 2. İş Mantığı: Son ping zamanını Redis'e yaz
	if err := s.cacheRepo.SetLastPing(ctx, now); err != nil {
		return err
	}

	return nil
}

func (s *pongService) CheckHealth(ctx context.Context) bool {
	return s.logRepo.Ping() == nil && s.cacheRepo.Ping(ctx) == nil
}
