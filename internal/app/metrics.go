package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go-metrics-collector/internal/config"
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/middleware"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/router"
	"go-metrics-collector/internal/service"
	"go-metrics-collector/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	cfg     config.ServerConfig
	router  *gin.Engine
	service service.MetricsService
	storage *storage.FileStorage
	logger  *zap.Logger
	pool    *pgxpool.Pool
	// нужно ли читать метрики из файла на старте и сбрасывать их туда
	storeEnabled bool
}

func NewMetricsApp(ctx context.Context, cfg config.ServerConfig) (*App, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать zap логер: %w", err)
	}

	var pool *pgxpool.Pool
	var metricsRepository repository.MetricsRepository

	if cfg.DatabaseDSN != "" {
		pool, err = pgxpool.New(ctx, cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("не удалось создать пул подключений: %w", err)
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
		}

		if err := repository.Migrate(cfg.DatabaseDSN); err != nil {
			pool.Close()
			return nil, fmt.Errorf("не удалось применить миграции: %w", err)
		}
	}

	if pool != nil {
		metricsRepository = repository.NewPgMetricsRepository(pool)
	} else {
		metricsRepository = repository.NewMetricsRepository()
	}

	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	dbService := service.NewDBService(repository.NewDBRepository(pool))
	dbHandler := handler.NewDBHandler(dbService)

	fileStorage := storage.NewFileStorage(cfg.FileStoragePath)

	// когда есть бд, то дампить и восстанавливать из файла не нужно
	storeEnabled := cfg.FileStoragePath != "" && pool == nil

	if storeEnabled && cfg.Restore {
		restoreMetrics(ctx, metricsService, fileStorage, logger)
	}

	var updateMiddlewares []gin.HandlerFunc
	if storeEnabled && cfg.StoreInterval == 0 {
		dump := func(ctx context.Context) error {
			return dumpMetrics(ctx, metricsService, fileStorage)
		}

		updateMiddlewares = append(updateMiddlewares, middleware.DumpMetrics(dump, logger))
	}

	metricsRouter := router.New(
		router.Handlers{
			Metrics: metricsHandler,
			DB:      dbHandler,
		},
		logger,
		updateMiddlewares...,
	)

	return &App{
		cfg:          cfg,
		router:       metricsRouter,
		service:      metricsService,
		storage:      fileStorage,
		logger:       logger,
		pool:         pool,
		storeEnabled: storeEnabled,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.logger.Sync()
	if a.pool != nil {
		defer a.pool.Close()
	}

	server := &http.Server{
		Addr:    a.cfg.Address,
		Handler: a.router,
	}
	var wg sync.WaitGroup

	if a.storeEnabled && a.cfg.StoreInterval > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.runMetricsDump(ctx)
		}()
	}

	serverErr := make(chan error, 1)

	go func() {
		a.logger.Info(
			"сервер запущен",
			zap.String("address", a.cfg.Address),
		)

		err := server.ListenAndServe()

		if !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		a.logger.Error("сервер завершился с ошибкой", zap.Error(err))
	case <-ctx.Done():
		a.logger.Info("завершение работы сервера...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("ошибка graceful shutdown", zap.Error(err))
	}

	wg.Wait()
	a.logger.Info("сервер остановлен")
	return nil
}

func (a *App) runMetricsDump(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := dumpMetrics(ctx, a.service, a.storage); err != nil {
				a.logger.Error("не удалось сохранить метрики на диск", zap.Error(err))
			}
		case <-ctx.Done():
			dumpCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := dumpMetrics(dumpCtx, a.service, a.storage); err != nil {
				a.logger.Error("не удалось сохранить метрики на диск", zap.Error(err))
			}
			return
		}
	}
}

func dumpMetrics(
	ctx context.Context,
	metricsService service.MetricsService,
	fileStorage *storage.FileStorage,
) error {
	metrics, err := metricsService.GetMetrics(ctx)
	if err != nil {
		return fmt.Errorf("не удалось прочитать метрики: %w", err)
	}

	return fileStorage.Dump(metrics)
}

func restoreMetrics(
	ctx context.Context,
	metricsService service.MetricsService,
	fileStorage *storage.FileStorage,
	logger *zap.Logger,
) {
	metrics, err := fileStorage.Load()
	if err != nil {
		logger.Warn("не удалось прочитать сохраненные метрики", zap.Error(err))
		return
	}

	if len(metrics) == 0 {
		logger.Info("сохранённых метрик нет")
		return
	}

	if err := metricsService.Restore(ctx, metrics); err != nil {
		logger.Warn("не удалось восстановить метрики", zap.Error(err))
		return
	}

	logger.Info("метрики восстановлены из файла")
}
