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
	rt.logger.Debug("register ldap router")
	v1 := r.Group("/api/v1/:cluster/ldap")
	{
		v1.GET("/user/list", rt.HandlerGetUserlist)      // GET /api/v1/:cluster/ldap/user/list
		v1.POST("/user", rt.HandlerPostUser)             // POST /api/v1/:cluster/ldap/user
		v1.PUT("/user/:name", rt.HandlerPutUser)         // PUT /api/v1/:cluster/ldap/user/:name
		v1.DELETE("/user/:name", rt.HandlerDeleteUser)   // DELETE /api/v1/:cluster/ldap/user/:name
		v1.GET("/group/list", rt.HandlerGetGroupList)    // GET /api/v1/:cluster/ldap/group/list
		v1.POST("/group", rt.HandlerAddGroup)            // POST /api/v1/:cluster/ldap/group
		v1.PUT("/group/:name", rt.HandlerUpdateGroup)    // PUT /api/v1/:cluster/ldap/group/:name
		v1.DELETE("/group/:name", rt.HandlerDeleteGroup) // DELETE /api/v1/:cluster/ldap/group/:name
	}
}
