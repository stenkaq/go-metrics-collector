package service

import (
	"context"

	"go-metrics-collector/internal/repository"
)

type DBService interface {
	Ping(ctx context.Context) error
}

type dbService struct {
	repository repository.Pinger
}

func NewDBService(r repository.Pinger) DBService {
	return &dbService{repository: r}
}

func (s *dbService) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}
