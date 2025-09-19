package ldap

import (
	"csjk-bk/internal/pkg/client/postgres"
	"csjk-bk/internal/pkg/client/slurmrest"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Router struct {
	db         *postgres.Client
	slurmrestc *slurmrest.Client
	logger     *slog.Logger
}

func NewRouter(db *postgres.Client, slurmrestc *slurmrest.Client, logger *slog.Logger) *Router {
	return &Router{
		db:         db,
		slurmrestc: slurmrestc,
		logger:     logger,
	}
}

func (rt *Router) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1/ldap")
	{
		v1.GET("/user/list", rt.HandlerGetUserlist) // GET /api/v1/:cluster/ldap/user/list
		v1.POST("/user")
		v1.PUT("/user/:name")
		v1.DELETE("/user/:name")
		v1.GET("/group/list")
		v1.POST("/group")
		v1.PUT("/group/:name")
		v1.DELETE("/group/:name")
	}
}
