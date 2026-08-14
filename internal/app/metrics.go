package app

import (
	"fmt"
	"time"

	"go-metrics-collector/internal/config"
	"go-metrics-collector/internal/handler"
	"go-metrics-collector/internal/middleware"
	"go-metrics-collector/internal/repository"
	"go-metrics-collector/internal/router"
	"go-metrics-collector/internal/service"
	"go-metrics-collector/internal/storage"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	cfg     config.ServerConfig
	router  *gin.Engine
	service service.MetricsService
	storage *storage.FileStorage
	logger  *zap.Logger
}

func NewMetricsApp(cfg config.ServerConfig) (*App, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("не удалось создать zap логер: %w", err)
	}

	metricsRepository := repository.NewMetricsRepository()
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)
	fileStorage := storage.NewFileStorage(cfg.FileStoragePath)

	storeEnabled := cfg.FileStoragePath != ""

	if storeEnabled && cfg.Restore {
		restoreMetrics(metricsService, fileStorage, logger)
	}

	var updateMiddlewares []gin.HandlerFunc
	if storeEnabled && cfg.StoreInterval == 0 {
		dump := func() error {
			return dumpMetrics(metricsService, fileStorage)
		}

		updateMiddlewares = append(updateMiddlewares, middleware.DumpMetrics(dump, logger))
	}

	metricsRouter := router.New(
		router.Handlers{
			Metrics: metricsHandler,
		},
		logger,
		updateMiddlewares...,
	)

	return &App{
		cfg:     cfg,
		router:  metricsRouter,
		service: metricsService,
		storage: fileStorage,
		logger:  logger,
	}, nil
}

func (a *App) Run() error {
	defer a.logger.Sync()

	if a.cfg.FileStoragePath != "" && a.cfg.StoreInterval > 0 {
		go a.runMetricsDump()
	}

	return a.router.Run(a.cfg.Address)
}

func (a *App) runMetricsDump() {
	ticker := time.NewTicker(a.cfg.StoreInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := dumpMetrics(a.service, a.storage); err != nil {
			a.logger.Error("не удалось сохранить метрики на диск", zap.Error(err))
		}
	}
}

func dumpMetrics(metricsService service.MetricsService, fileStorage *storage.FileStorage) error {
	return fileStorage.Dump(metricsService.GetMetrics())
}

func restoreMetrics(
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

	metricsService.Restore(metrics)
	logger.Info("метрики восстановлены из файла")
}
