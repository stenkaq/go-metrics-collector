package app

import (
	"fmt"
	"time"

	"go-metrics-collector/internal/config"
	"go-metrics-collector/internal/handler"
	models "go-metrics-collector/internal/model"
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

	metricsRouter := router.New(
		router.Handlers{
			Metrics: metricsHandler,
		},
		logger,
	)

	return &App{
		cfg:     cfg,
		router:  metricsRouter,
		service: metricsService,
		storage: storage.NewFileStorage(cfg.FileStoragePath),
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
		if err := a.metricsDump(); err != nil {
			a.logger.Error("не удалось сохранить метрики на диск", zap.Error(err))
		}
	}
}

func (a *App) metricsDump() error {
	metrics := a.service.GetMetrics()

	list := make([]models.Metrics, 0, len(metrics))
	for _, metric := range metrics {
		list = append(list, metric)
	}

	return a.storage.Dump(list)
}
