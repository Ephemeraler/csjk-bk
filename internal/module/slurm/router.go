package slurm

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
	v1 := r.Group("/api/v1/")
	{
		g := v1.Group("/:cluster/slurm")
		g.GET("/overview", rt.HandlerGetOverview)                                              // GET /api/v1/:cluster/slurm/overview
		g.GET("/qos/list", rt.HandlerGetQoSList)                                               // GET /api/v1/:cluster/slurm/qos/list?paging=xxx&page=xxx&page_size=xxx
		g.GET("qos/name/list", rt.HandlerGetQosNameList)                                       // GET /api/v1/:cluster/slurm/qos/name/list
		g.GET("/qos/:id/detail", rt.HandlerGetQoSDetail)                                       // GET /api/v1/:cluster/slurm/qos/:id/detail
		g.GET("/account/name/list", rt.HandlerGetAccountsNameList)                             // GET /api/v1/:cluster/slurm/account/name/list
		g.GET("/account/:account/childnodes", rt.HandlerGetAccountChildNodes)                  // GET /api/v1/:cluster/slurm//account/:account/childnodes
		g.GET("/association/:account/childnodes", rt.HandlerGetAssociationChildNodesOfAccount) // GET /api/v1/:cluster/slurm/association/:account/childnodes
		g.GET("/association/detail", rt.HandlerGetAssociationDetail)                           // GET /api/v1/:cluster/slurm/association/detail?account=xxx&user=xxx&partition=xxx
		g.GET("/accounting/job/list", rt.HandlerGetJobListFromAccounting)                      // GET /api/v1/:cluster/slurm/accounting//job/list?paging=xxx&page=xxx&page_size=xxx
		g.GET("/accounting/job/:jobid/detail", rt.HandlerGetAccountingJobDetail)               // GET /api/v1/:cluster/slurm/accounting/job/:jobid/detail
		g.GET("/partition/list", rt.HandlerGetPartitionList)                                   // GET /api/v1/:cluster/slurm/partition/list?paging=xxx&page=xxx&page_size=xxx
		g.GET("partition/:name/detail", rt.HandlerGetPartitionDetail)                          // GET /api/v1/:cluster/slurm/partition/:name/detail
		g.GET("/scheduling/job/list", rt.HandlerGetSchedulingJobList)                          // GET /api/v1/:cluster/slurm/scheduling/job/list
		g.GET("/scheduling/job/:jobid/detail", rt.HandlerGetSchedulingJobsDetail)              // GET /api/v1/:cluster/slurm/scheduling/job/:jobid/detail
	}
}
