package handler

import (
	"net/http"

	"go-metrics-collector/internal/service"

	"github.com/gin-gonic/gin"
)

type DBHandler interface {
	Ping(c *gin.Context)
}

type dbHandler struct {
	service service.DBService
}

func NewDBHandler(s service.DBService) DBHandler {
	return &dbHandler{service: s}
}

func (h *dbHandler) Ping(c *gin.Context) {
	if err := h.service.Ping(c.Request.Context()); err != nil {
		http.Error(c.Writer, "Ошибка при пинге БД", http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
