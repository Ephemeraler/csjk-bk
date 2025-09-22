package lustre

import (
	"csjk-bk/internal/pkg/client/lustre"
	"csjk-bk/internal/pkg/client/postgres"
	"csjk-bk/internal/pkg/client/slurmrest"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Router struct {
	db           *postgres.Client
	slurmrestc   *slurmrest.Client
	lustreClient *lustre.Client
	logger       *slog.Logger
}

func NewRouter(db *postgres.Client, slurmrestc *slurmrest.Client, lc *lustre.Client, logger *slog.Logger) *Router {
	return &Router{db: db, slurmrestc: slurmrestc, lustreClient: lc, logger: logger}
}

func (rt *Router) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1/:cluster/lustre")
	{
		v1.GET("/quotas", rt.HandlerGetQuotas)                                   // GET /api/v1/:cluster/lustre/quotas?user=xxx&userxxx&paging=xxx&page=xxx&page_size=xxx
		v1.PUT("/:user/quota", rt.HandlerUpdateUserQuota)                        // PUT /api/v1/:cluster/lustre/:user/quota
		v1.PUT("/quota", rt.HandlerUpdateQuota)                                  // PUT /api/v1/:cluster/lustre/quota
		v1.GET("/quota/applications", rt.HandlerGetQuotaApps)                    // GET /api/v1/:cluster/lustre/quota/applications?user_id=xxx
		v1.GET("/quota/application/:id/decision", rt.HandlerGetQuotaAppDecision) // GET /api/v1/:cluster/lustre/quota/application/:id/decision
		v1.POST("/quota/application", rt.HandlerCreateQuotaApplication)          // POST /api/v1/:cluster/lustre/quota/application/:id
		v1.PUT("/quota/application/:id", rt.HandlerUpdateQuotaApplication)       // PUT /api/v1/:cluster/lustre/quota/application/:id
		v1.DELETE("/quota/application/:id", rt.HandleDelQuotaApplication)        // DELETE /api/v1/:cluster/lustre/quota/application/:id
		v1.POST("/quota/application/:id/review", rt.HandlerPostQuotaAppReview)   // POST /api/v1/:cluster/lustre/quota/application/:id/review
	}
}
