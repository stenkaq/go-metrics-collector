package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBRepository interface {
	Ping(ctx context.Context) error
}

type dbRepository struct {
	pool *pgxpool.Pool
}

func NewDBRepository(pool *pgxpool.Pool) DBRepository {
	return &dbRepository{pool: pool}
}

func (r *dbRepository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return errors.New("подключение к БД отсутствует")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return r.pool.Ping(ctx)
}
