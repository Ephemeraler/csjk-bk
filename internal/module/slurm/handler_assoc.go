package slurm

import (
	"net/http"
	"strconv"
	"strings"

	"csjk-bk/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// @Summary 获取某集群中指定账户的关联子节点信息
// @Description 返回账户关联树中该账户的子账户与子用户节点信息
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param account path string true "账户名称" example("root")
// @Success 200 {object} response.Response{results=slurmrest.AssociationNode}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/association/{account}/childnodes [get]
func (rt *Router) HandlerGetAssociationChildNodesOfAccount(c *gin.Context) {
	// 校验路径参数
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}
	account := c.Param("account")
	if account == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing account in path"})
		return
	}

	// 获取 slurmrestd 地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 查询关联子节点
	node, err := rt.slurmrestc.GetAssociationChildNodesOfAccount(c.Request.Context(), addr, account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch association child nodes: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: node})
}

type AssociationDetail struct {
	IDAssoc        uint32 `json:"id_assoc"`          // 关联 ID
	ClusterName    string `json:"cluster_name"`      // 系统名称
	Acct           string `json:"acct"`              // 账号名称
	Partition      string `json:"partition"`         // 分区名称
	Shares         int32  `json:"shares"`            // 公平份额权重
	MaxJobs        int32  `json:"max_jobs"`          // 单账户运行作业上限
	MaxSubmitJobs  int32  `json:"max_submit_jobs"`   // 单账户提交作业上限
	MaxWallPJ      int32  `json:"max_wall_pj"`       // 单作业最大运行时间
	GrpWall        int32  `json:"grp_wall"`          // 组级总运行时间限制
	GrpTres        string `json:"grp_tres"`          // 组级总 TRES 资源限制
	GrpTresMins    string `json:"grp_tres_mins"`     // 组级 TRES 时间限制
	GrpJobs        int32  `json:"grp_jobs"`          // 组级运行作业总数上限
	GrpSubmitJobs  int32  `json:"grp_submit_jobs"`   // 组级提交作业总数上限
	Priority       uint32 `json:"priority"`          // 账户调度优先级
	MaxJobsAccrue  int32  `json:"max_jobs_accrue"`   // 累计优先级作业上限
	MinPrioThresh  int32  `json:"min_prio_thresh"`   // 优先级阈值
	GrpJobsAccrue  int32  `json:"grp_jobs_accrue"`   // 组级累计优先级作业上限
	MaxTresPJ      string `json:"max_tres_pj"`       // 单作业 TRES 上限
	MaxTresMinsPJ  string `json:"max_tres_mins_pj"`  // 单作业 TRES 时间上限
	GrpTresRunMins string `json:"grp_tres_run_mins"` // 组级运行中 TRES 时间上限
	MaxTresPN      string `json:"max_tres_pn"`       // 单节点 TRES 上限
	MaxTresRunMins string `json:"max_tres_run_mins"` // 单作业运行中 TRES 时间上限
	DefQosID       string `json:"def_qos_id"`        // 默认 QoS 策略
	QOS            string `json:"qos"`               // 可用/关联的 QoS 列表
}

// @Summary 获取某集群中某个关联详情
// @Description 根据 account、可选 user、可选 partition 查询关联详情
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param account query string true "账户名" example("root")
// @Param user query string false "用户名" default("") example("alice")
// @Param partition query string false "分区名" default("") example("p1")
// @Success 200 {object} response.Response{results=AssociationDetail}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/association/detail [get]
func (rt *Router) HandlerGetAssociationDetail(c *gin.Context) {
	// 校验 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析查询参数
	acct := c.Query("account")
	if acct == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing required query: account"})
		return
	}
	user := c.Query("user")
	part := c.Query("partition")

	// 解析 slurmrestd 地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 查询详情
	item, err := rt.slurmrestc.GetAssociationDetail(c.Request.Context(), addr, acct, user, part)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch association detail: " + err.Error()})
		return
	}

	qosItems, _, err := rt.slurmrestc.GetQosAll(c.Request.Context(), addr, false, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch qos list: " + err.Error()})
		return
	}

	idToName := make(map[int32]string)
	for _, qos := range qosItems {
		idToName[qos.ID] = qos.Name
	}

	qosList := make([]string, 0)
	for _, idstr := range strings.Split(item.QOS, ",") {
		id, err := strconv.Atoi(idstr)
		if err != nil {
			continue
		}
		if n, ok := idToName[int32(id)]; ok {
			qosList = append(qosList, n)
		}

	}

	// 映射输出
	out := AssociationDetail{
		IDAssoc:        item.IDAssoc,
		ClusterName:    "", // slurmrestd 未返回 cluster 字段；如需可从路径 cluster 注入
		Acct:           item.Acct,
		Partition:      item.Partition,
		Shares:         item.Shares,
		MaxJobs:        item.MaxJobs,
		MaxSubmitJobs:  item.MaxSubmitJobs,
		MaxWallPJ:      item.MaxWallPJ,
		GrpWall:        item.GrpWall,
		GrpTres:        item.GrpTres,
		GrpTresMins:    item.GrpTresMins,
		GrpJobs:        item.GrpJobs,
		GrpSubmitJobs:  item.GrpSubmitJobs,
		Priority:       item.Priority,
		MaxJobsAccrue:  item.MaxJobsAccrue,
		MinPrioThresh:  item.MinPrioThresh,
		GrpJobsAccrue:  item.GrpJobsAccrue,
		MaxTresPJ:      item.MaxTresPJ,
		MaxTresMinsPJ:  item.MaxTresMinsPJ,
		GrpTresRunMins: item.GrpTresRunMins,
		MaxTresPN:      item.MaxTresPN,
		MaxTresRunMins: item.MaxTresRunMins,
		DefQosID:       idToName[item.DefQosID],
		QOS:            strings.Join(qosList, ","),
	}
	// 如需要返回 cluster，可 out.ClusterName = cluster

	c.JSON(http.StatusOK, response.Response{Results: out})
}
