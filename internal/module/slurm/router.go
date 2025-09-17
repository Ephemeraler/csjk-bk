package slurm

import (
	"csjk-bk/internal/pkg/client/postgres"
	"csjk-bk/internal/pkg/client/slurmrest"

	"github.com/gin-gonic/gin"
)

type Router struct {
	db         *postgres.Client
	slurmrestc *slurmrest.Client
}

func NewRouter(db *postgres.Client, slurmrestc *slurmrest.Client) *Router {
	return &Router{
		db:         db,
		slurmrestc: slurmrestc,
	}
}

func (rt *Router) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1/")
	{
		g := v1.Group("/:cluster/slurm")
		g.GET("/overview", rt.HandlerGetOverview)        // GET /api/v1/{cluster}/slurm/overview
		g.GET("/qos/list", rt.HandlerGetQoSList)         // GET /api/v1/{cluster}/slurm/qos/list
		g.GET("/qos/:id/detail", rt.HandlerGetQoSDetail) // GET /api/v1/{cluster}/slurm/qos/{id}/detail
	}
}
