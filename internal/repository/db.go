package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	models "go-metrics-collector/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type dbRepository struct {
	pool *pgxpool.Pool
}

func NewDBRepository(pool *pgxpool.Pool) Pinger {
	return &dbRepository{pool: pool}
}

func (p *dbRepository) Ping(ctx context.Context) error {
	if p.pool == nil {
		return errors.New("подключение к БД отсутствует")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.pool.Ping(ctx)
}

type pgMetricsRepository struct {
	pool *pgxpool.Pool
}

func NewPgMetricsRepository(pool *pgxpool.Pool) MetricsRepository {
	return &pgMetricsRepository{pool: pool}
}

func (r *pgMetricsRepository) Save(ctx context.Context, metric models.Metrics) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id, type) DO UPDATE
		SET delta = EXCLUDED.delta,
	    value = EXCLUDED.value`,
		metric.ID, metric.MType, metric.Delta, metric.Value)
	if err != nil {
		return fmt.Errorf("не удалось сохранить метрику %q: %w", metric.ID, err)
	}

	return nil
}

func (r *pgMetricsRepository) IncrementCounter(ctx context.Context, name string, delta int64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO metrics (id, type, delta, value)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (id, type) DO UPDATE
		SET delta = metrics.delta + EXCLUDED.delta`, name, models.Counter, delta)
	if err != nil {
		return fmt.Errorf("не удалось увеличить счётчик %q: %w", name, err)
	}

	return nil
}

func (r *pgMetricsRepository) Get(ctx context.Context, mType string, name string) (models.Metrics, bool, error) {
	var metric models.Metrics

	err := r.pool.QueryRow(ctx,
		`SELECT id, type, delta, value
		FROM metrics
		WHERE id = $1 AND type = $2`, name, mType).
		Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Metrics{}, false, nil
	}

	if err != nil {
		return models.Metrics{}, false, fmt.Errorf("не удалось прочитать метрику %q: %w", name, err)
	}

	return metric, true, nil
}

func (r *pgMetricsRepository) GetMetrics(ctx context.Context) ([]models.Metrics, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, type, delta, value FROM metrics`)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать метрики: %w", err)
	}
	defer rows.Close()

	metrics := make([]models.Metrics, 0)

	for rows.Next() {
		var metric models.Metrics

		if err := rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value); err != nil {
			return nil, fmt.Errorf("не удалось разобрать строку с метрикой: %w", err)
		}

		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("не удалось дочитать метрики: %w", err)
	}

	return metrics, nil
}
