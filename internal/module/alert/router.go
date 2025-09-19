package alert

import (
	"csjk-bk/internal/pkg/client/alertmanager"
	"csjk-bk/internal/pkg/client/postgres"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Router struct {
	db       *postgres.Client
	amClient *alertmanager.Client
	logger   *slog.Logger
}

func NewRouter(db *postgres.Client, amClient *alertmanager.Client, logger *slog.Logger) *Router {
	return &Router{
		db:       db,
		amClient: amClient,
		logger:   logger,
	}
}

func (rt *Router) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1/alerts")
	{
		v1.GET("/firing", rt.HandlerGetAlertsFiring)
		v1.GET("/history", rt.HandlerGetAlertsHistory)
	}
}
