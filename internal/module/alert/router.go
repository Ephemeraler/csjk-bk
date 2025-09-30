package alert

import (
	"csjk-bk/internal/pkg/client/alertmanager"
	"csjk-bk/internal/pkg/client/exec"
	"csjk-bk/internal/pkg/client/postgres"
	"log/slog"
	osexec "os/exec"

	"github.com/gin-gonic/gin"
)

type Router struct {
	db         *postgres.Client
	amClient   *alertmanager.Client
	execClient *exec.Client
	logger     *slog.Logger
}

func NewRouter(db *postgres.Client, amClient *alertmanager.Client, logger *slog.Logger) *Router {
	execClient := &exec.Client{}
	execClient.Set(osexec.CommandContext, logger)
	return &Router{
		db:         db,
		amClient:   amClient,
		execClient: execClient,
		logger:     logger,
	}
}

func (rt *Router) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1/alerts")
	{
		v1.GET("/firing", rt.HandlerGetAlertsFiring)
		v1.GET("/history", rt.HandlerGetAlertsHistory)
		v1.POST("/outband/setting/sensor/upper/thresholds", rt.HandlerSetUpperThreshsOfOutbandSensor)
		v1.POST("/outband/setting/sensor/lower/thresholds", rt.HandlerSetLowerThreshsOfOutbandSensor)
		v1.POST("/outband/setting/sensor/threshold", rt.HandlerSetThreshOfOutbandSensor)
		v1.POST("/outband/setting/sensor/inhibit", rt.HandlerSetInhibitOfOutbandSensor)
		v1.GET("/outband/sensor/thresholds", rt.HandlerGetThreshOfOutbandSensor)
	}
}
