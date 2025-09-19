package slurm

import (
	"net/http"
	"net/url"
	"strconv"

	"csjk-bk/internal/pkg/common/paging"
	"csjk-bk/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type QoSList []QoSListElem
type QoSListElem struct {
	ID                      int32  `json:"id"`                              // ID
	Name                    string `json:"Name"`                            // 名称
	Description             string `json:"Description"`                     // 描述
	MaxJobsPerAccount       int32  `json:"max_jobs_per_account"`            // 每账号作业数限制
	MaxJobsPerUser          int32  `json:"max_jobs_per_user"`               // 每⽤⼾作业数限制
	MaxSubmitJobsPerAccount int32  `json:"max_submit_jobs_per_account"`     // 每账号提交作业数限制
	MaxSubmitJobsPerUser    int32  `json:"max_submit_jobs_per_user"`        // 每⽤⼾提交作业数限制
	MaxWallDurationPerJob   int32  `json:"max_wall_duration_per_job"`       // 作业运⾏时间限制
	GrpJobs                 int32  `json:"grp_jobs"`                        // 总作业数限制
	GrpSubmitJobs           int32  `json:"grp_submit_jobs"`                 // 总提交作业数限制
	GrpWall                 int32  `gorm:"column:grp_wall" json:"grp_wall"` // 总运⾏时间限制
}

// @Summary 获取某集群中 QoS 列表
// @Tags 资源管理, QoS
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param paging query bool false "是否开启分页" default(false)
// @Param page query int false "页码，从 1 开始（仅当 paging=true 生效）" minimum(1) default(1)
// @Param page_size query int false "每页数量，1-100（仅当 paging=true 生效）" minimum(1) maximum(100) default(20)
// @Success 200 {object} response.Response{results=QoSList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/qos/list [get]
func (rt *Router) HandlerGetQoSList(c *gin.Context) {
	// 解析路径参数 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	pq.SetDefaults(1, 20, 100)

	// 查询 slurmrestd 地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 调用 slurmrest 客户端，获取当前页数据
	items, total, err := rt.slurmrestc.GetQosAll(c.Request.Context(), addr, pq.Paging, pq.Page, pq.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch qos list: " + err.Error()})
		return
	}

	// 获取总数用于分页：若分页开启，则再拉取一次不分页数据以获取总量
	var prev, next url.URL
	if pq.Paging {
		prev, next = response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
	}

	// 组装输出列表
	out := make(QoSList, 0, len(items))
	for _, it := range items {
		out = append(out, QoSListElem{
			ID:                      it.ID,
			Name:                    it.Name,
			Description:             it.Description,
			MaxJobsPerAccount:       it.MaxJobsPA,
			MaxJobsPerUser:          it.MaxJobsPerUser,
			MaxSubmitJobsPerAccount: it.MaxSubmitJobsPA,
			MaxSubmitJobsPerUser:    it.MaxSubmitJobsPerUser,
			MaxWallDurationPerJob:   it.MaxWallDurationPerJob,
			GrpJobs:                 it.GrpJobs,
			GrpSubmitJobs:           it.GrpSubmitJobs,
			GrpWall:                 it.GrpWall,
		})
	}
	c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: out})
}

type QoSDetail struct {
	ID                    int32   `json:"id"`                        // ID
	Name                  string  `json:"name"`                      // 名称
	Description           string  `json:"description"`               // 描述
	GraceTime             uint32  `json:"grace_time"`                // 宽限时间
	MaxJobsPerAccount     int32   `json:"max_jobs_per_account"`      // 每账号作业数限制
	MaxJobsPerUser        int32   `json:"max_jobs_per_user"`         // 每⽤⼾作业数限制
	MaxJobsAccruePA       int32   `json:"max_jobs_accrue_pa"`        // 每账号累计优先级作业数限制
	MaxJobsAccruePU       int32   `json:"max_jobs_accrue_pu"`        // 每用户累计优先级作业数限制
	GrpJobs               int32   `json:"grp_jobs"`                  // 总作业数限制
	GrpJobsAccrue         int32   `json:"grp_jobs_accrue"`           // 总累计优先级作业数限制
	MaxSubmitJobsPA       int32   `json:"max_submit_jobs_pa"`        // 每账号提交作业数限制
	MaxSubmitJobsPerUser  int32   `json:"max_submit_jobs_per_user"`  // 每用户提交作业限制
	GrpSubmitJobs         int32   `json:"grp_submit_jobs"`           // 总提交作业限制
	MaxWallDurationPerJob int32   `json:"max_wall_duration_per_job"` // 作业运行时间限制
	MaxTresRunMinsPA      string  `json:"max_tres_run_mins_pa"`      // 账号TRES运行时间限制
	GrpWall               int32   `json:"grp_wall"`                  // 总运行时间限制
	GrpTresRunMins        string  `json:"grp_tres_run_mins"`         // 总TRES运行时间限制
	MaxTresMinsPJ         string  `json:"max_tres_mins_pj"`          // 作业TRES时间限制
	MaxTresRunMinsPU      string  `json:"max_tres_run_mins_pu"`      // 用户TRES运行时间限制
	GrpTresMins           string  `json:"grp_tres_mins"`             // 总TRES时间限制
	MaxTresPA             string  `json:"max_tres_pa"`               // 账号TRES限制
	MaxTresPN             string  `json:"max_tres_pn"`               // 节点TRES限制
	MinTresPJ             string  `json:"min_tres_pj"`               // 作业TRES下限
	MaxTresPJ             string  `json:"max_tres_pj"`               // 作业TRES限制
	MaxTresPU             string  `json:"max_tres_pu"`               // 用户TRES限制
	GrpTres               string  `json:"grp_tres"`                  // 总TRES限制
	MinPrioThresh         int32   `json:"min_prio_thresh"`           // 优先级阈值
	Preempt               string  `json:"preempt"`                   // 可被抢占的QoS
	PreemptExemptTime     uint32  `json:"preempt_exempt_time"`       // 抢占豁免时间
	Priority              uint32  `json:"priority"`                  // 优先级因子
	PreemptMode           int32   `json:"preempt_mode"`              // 抢占模式
	UsageFactor           float64 `json:"usage_factor"`              // 资源使用因子
	UsageThres            float64 `json:"usage_thres"`               // 资源使用阈值
	LimitFactor           float64 `json:"limit_factor"`              // 资源限制因子
}

// @Summary 获取某集群中某个 QoS 详情
// @Tags 资源管理, QoS
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param id path int true "QoS ID"
// @Success 200 {object} response.Response{results=QoSDetail}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/qos/{id}/detail [get]
func (rt *Router) HandlerGetQoSDetail(c *gin.Context) {
	// 解析路径参数 cluster 和 id
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid path param: id"})
		return
	}

	// 查询 slurmrestd 地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 调用 slurmrest 客户端
	q, err := rt.slurmrestc.GetQos(c.Request.Context(), addr, uint32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch qos detail: " + err.Error()})
		return
	}

	out := QoSDetail{
		ID:                    q.ID,
		Name:                  q.Name,
		Description:           q.Description,
		GraceTime:             q.GraceTime,
		MaxJobsPerAccount:     q.MaxJobsPA,
		MaxJobsPerUser:        q.MaxJobsPerUser,
		MaxJobsAccruePA:       q.MaxJobsAccruePA,
		MaxJobsAccruePU:       q.MaxJobsAccruePU,
		GrpJobs:               q.GrpJobs,
		GrpJobsAccrue:         q.GrpJobsAccrue,
		MaxSubmitJobsPA:       q.MaxSubmitJobsPA,
		MaxSubmitJobsPerUser:  q.MaxSubmitJobsPerUser,
		GrpSubmitJobs:         q.GrpSubmitJobs,
		MaxWallDurationPerJob: q.MaxWallDurationPerJob,
		MaxTresRunMinsPA:      q.MaxTresRunMinsPA,
		GrpWall:               q.GrpWall,
		GrpTresRunMins:        q.GrpTresRunMins,
		MaxTresMinsPJ:         q.MaxTresMinsPJ,
		MaxTresRunMinsPU:      q.MaxTresRunMinsPU,
		GrpTresMins:           q.GrpTresMins,
		MaxTresPA:             q.MaxTresPA,
		MaxTresPN:             q.MaxTresPN,
		MinTresPJ:             q.MinTresPJ,
		MaxTresPJ:             q.MaxTresPJ,
		MaxTresPU:             q.MaxTresPU,
		GrpTres:               q.GrpTres,
		MinPrioThresh:         q.MinPrioThresh,
		Preempt:               q.Preempt,
		PreemptExemptTime:     q.PreemptExemptTime,
		Priority:              q.Priority,
		PreemptMode:           q.PreemptMode,
		UsageFactor:           q.UsageFactor,
		UsageThres:            q.UsageThres,
		// LimitFactor 不存在于模型中，保持为零值
	}

	c.JSON(http.StatusOK, response.Response{Results: out})
}

type QoSNameList []string

// @Summary 获取某集群中所有 QoS 名称列表
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=QoSNameList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/qos/name/list [get]
func (rt *Router) HandlerGetQosNameList(c *gin.Context) {
	// 解析路径参数 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 查询 slurmrestd 地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 获取全部 QoS 列表（不分页）
	items, total, err := rt.slurmrestc.GetQosAll(c.Request.Context(), addr, false, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch qos list: " + err.Error()})
		return
	}

	// 提取名称列表
	names := make(QoSNameList, 0, len(items))
	for _, q := range items {
		if q.Name != "" {
			names = append(names, q.Name)
		}
	}

	c.JSON(http.StatusOK, response.Response{Count: total, Results: names})
}
