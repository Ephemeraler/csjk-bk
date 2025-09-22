package lustre

import (
	"csjk-bk/internal/pkg/common/paging"
	"csjk-bk/internal/pkg/common/time"
	"csjk-bk/internal/pkg/response"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	dbpg "csjk-bk/internal/pkg/client/postgres"

	"github.com/gin-gonic/gin"
)

// TODO: KD 响应结果中的 Key, 暂时还未确定, 后续根据情况直接设置可用.
const (
	LUSTRE_BLOCK_QUOTA_SOFT_LIMIT = ""
	LUSTRE_BLOCK_QUOTA_HARD_LIMIT = ""
	LUSTRE_BLOCK_QUOTA_GRACE      = ""
	LUSTRE_FILE_QUOTA_SOFT_LIMIT  = ""
	LUSTRE_FILE_QUOTA_HARD_LIMIT  = ""
	LUSTRE_FILE_QUOTA_GRACE       = ""
	LUSTRE_FILESYSTEM             = ""
	LUSTRE_MOUNTED                = "Mounted"
)

type UserQuotaList []UserQuota
type UserQuota struct {
	User       string `json:"user"`                   // 用户
	FileSystem string `json:"filesystem"`             // 文件系统
	BQSL       string `json:"block_quota_soft_limit"` // 块配额软限制
	BQHL       string `json:"block_quota_hard_limit"` // 块配额硬限制
	BQG        string `json:"block_quota_grace"`      // 块配额宽限期
	FQSL       string `json:"file_quota_soft_limit"`  // 文件配额软限制
	FQHL       string `json:"file_quota_hard_limit"`  // 文件配额硬限制
	FQG        string `json:"file_quota_grace"`       // 文件配额宽限期
}

// HandlerGetQuotas 该 API 设计如下:
// api: GET /api/v1/:cluster/lustre/quotas?user=xxx&user=xxx&paging=xxx&page=xxx&page_size=xxx
// 执行流程:
//   - 解析参数
//   - 是否有 user 参数, 若没有则调用 slurmrest.GetLdapUsers 获取全部用户信息.
//   - 调用 lustreClient.GetMounts 函数获取lustre挂载点信息, 挂载点 key 为 LUSTRE_MOUNTED, 构建挂载点列表.
//   - 根据挂载点列表, 循环调用 lustreClient.GetDefaultQuota 获取每个lustre文件系统的默认配额.
//   - 根据挂载点列表、用户名, 循环调用 lustreClient.GetUserQuota 获取所有用户在所有文件系统的配额限制.
//     若没有查询到限制或限制的value为none, 则使用对应文件系统的默认配额限制值. 返回结果中对应 UserQuota 所需数据的 key 为 LUSTRE_*常量
//   - 根据分页参数决定是否分页及如何分页, 返回数据.
//
// @Summary 获取某集群 Lustre 用户配额
// @Description 获取某集群 Lustre 用户配额. 当未指定user时, 将会获取该集群所有用户(ldap)的配额. 参数 cluster 为保留参数, 当前整体设计就不支持多集群.
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param user query []string false "用户名，可重复传多个 user=alice&user=bob"
// @Param paging query bool false "是否分页" default(true)
// @Param page query int false "页码，从1开始" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Success 200 {object} response.Response{results=UserQuotaList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quotas [get]
func (rt *Router) HandlerGetQuotas(c *gin.Context) {
	// Summary: 获取某集群 Lustre 配额信息（支持按用户和分页）
	// Flow:
	// 1) 解析 cluster、分页与用户参数；根据 cluster 解析服务地址。
	// 2) 若未提供 user，则不分页拉取该集群的全部用户名。
	// 3) 调用 KD 接口获取 Lustre 挂载点，再获取每个挂载点的默认配额。
	// 4) 为每个用户在每个挂载点查询配额，若未设置或为 none，则回退为默认配额。
	// 5) 如开启分页，对结果进行分页并返回。

	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	addr, err := rt.db.GetLustreServer(cluster)
	if err != nil || strings.TrimSpace(addr) == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// Paging params
	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	pq.SetDefaults(1, 20, 100)

	// Users from query: user can appear multiple times
	users := c.QueryArray("user")
	if len(users) == 0 {
		// Fetch all users without paging
		addr, err := rt.db.GetSlurmrestdAddr(cluster)
		if err != nil || addr == "" {
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
			}
			return
		}
		items, _, err := rt.slurmrestc.GetLdapUsers(c.Request.Context(), addr, false, 0, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch ldap users: " + err.Error()})
			return
		}
		for _, m := range items {
			if u := firstNonEmptyLocal(m["uid"], m["name"], m["UID"], m["User"]); strings.TrimSpace(u) != "" {
				users = append(users, u)
			}
		}
	}

	// Get mounts and defaults
	mounts, err := rt.lustreClient.GetMounts(c.Request.Context(), addr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch lustre mounts: " + err.Error()})
		return
	}
	// Build mount list by key LUSTRE_MOUNTED
	mountPoints := make([]string, 0, len(mounts))
	for _, m := range mounts {
		if v := strings.TrimSpace(m[LUSTRE_MOUNTED]); v != "" {
			mountPoints = append(mountPoints, v)
		}
	}

	// Default quotas cache per mount
	defQuota := make(map[string]map[string]string)
	for _, mp := range mountPoints {
		dq, err := rt.lustreClient.GetDefaultQuota(c.Request.Context(), addr, mp)
		if err != nil {
			// 若单个挂载点失败，跳过该挂载点
			continue
		}
		defQuota[mp] = dq
	}

	// Collect user quotas
	list := make(UserQuotaList, 0, len(users)*max(1, len(mountPoints)))
	for _, u := range users {
		for _, mp := range mountPoints {
			uqMap, err := rt.lustreClient.GetUserQuota(c.Request.Context(), addr, u, mp)
			if err != nil {
				// 若查询失败，视为未设置，使用默认
				uqMap = map[string]string{}
			}
			dq := defQuota[mp]
			// Compose output with fallback to default when value is none/empty
			item := UserQuota{
				User:       u,
				FileSystem: firstNonEmptyLocal(uqMap[LUSTRE_FILESYSTEM], dq[LUSTRE_FILESYSTEM], mp),
				BQSL:       pickWithDefault(uqMap, dq, LUSTRE_BLOCK_QUOTA_SOFT_LIMIT),
				BQHL:       pickWithDefault(uqMap, dq, LUSTRE_BLOCK_QUOTA_HARD_LIMIT),
				BQG:        pickWithDefault(uqMap, dq, LUSTRE_BLOCK_QUOTA_GRACE),
				FQSL:       pickWithDefault(uqMap, dq, LUSTRE_FILE_QUOTA_SOFT_LIMIT),
				FQHL:       pickWithDefault(uqMap, dq, LUSTRE_FILE_QUOTA_HARD_LIMIT),
				FQG:        pickWithDefault(uqMap, dq, LUSTRE_FILE_QUOTA_GRACE),
			}
			list = append(list, item)
		}
	}

	total := len(list)
	if pq.Paging {
		start := (pq.Page - 1) * pq.PageSize
		if start < 0 {
			start = 0
		}
		end := start + pq.PageSize
		if start > total {
			list = list[:0]
		} else {
			if end > total {
				end = total
			}
			list = list[start:end]
		}
		prev, next := response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
		c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: list})
		return
	}

	c.JSON(http.StatusOK, response.Response{Count: total, Results: list})
}

// firstNonEmptyLocal returns the first non-empty string.
func firstNonEmptyLocal(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// pickWithDefault returns v[key] if non-empty and not "none", otherwise d[key].
func pickWithDefault(v, d map[string]string, key string) string {
	if v != nil {
		if s := strings.TrimSpace(v[key]); s != "" && !equalNone(s) {
			return s
		}
	}
	if d != nil {
		if s := strings.TrimSpace(d[key]); s != "" {
			return s
		}
	}
	return ""
}

func equalNone(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "none")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// HandlerUpdateUserQuota 更新用户配额.
// 执行流程:
//   - 从请求路径中获取 cluster, user 参数;
//   - 从请求体中获取更新用户配额信息 UserQuota;
//   - 构建 lustre 配额配置命令
//     lfs setquota -u <user> -b <block-softlimit> -B <block-hardlimit> -i <inode-softlimit> -I <inode-hardlimit> <挂载点, 即文件系统>
//     lfs setquota -t --block-grace=xxx --inode-grace=xxx <挂载点, 即文件系统>
//   - 调用 lustreClient.Control 执行上述两个命令.
//
// @Summary 在某集群更新用户 Lustre 配额
// @Description 在某集群更新用户 Lustre 配额.
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param user path string true "用户名"
// @Param body body UserQuota true "配额参数（需包含 filesystem，其他字段留空表示不更新）"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/:user/quota [put]
func (rt *Router) HandlerUpdateUserQuota(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析 user（路径参数）
	user := c.Param("user")
	if strings.TrimSpace(user) == "" {
		// 兼容可能的 :name
		user = c.Param("name")
	}
	if strings.TrimSpace(user) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing user in path"})
		return
	}

	// 解析 body
	var in UserQuota
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}
	mount := strings.TrimSpace(in.FileSystem)
	if mount == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "filesystem is required"})
		return
	}

	addr, err := rt.db.GetLustreServer(cluster)
	if err != nil || strings.TrimSpace(addr) == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 构建 setquota 命令（仅为非空且非 none 的字段添加参数）
	// lfs setquota -u <user> [-b BQSL] [-B BQHL] [-i FQSL] [-I FQHL] <mount>
	var parts []string
	parts = append(parts, "lfs", "setquota", "-u", fmt.Sprintf("%s", user))
	if s := strings.TrimSpace(in.BQSL); s != "" && !equalNone(s) {
		parts = append(parts, "-b", s)
	}
	if s := strings.TrimSpace(in.BQHL); s != "" && !equalNone(s) {
		parts = append(parts, "-B", s)
	}
	if s := strings.TrimSpace(in.FQSL); s != "" && !equalNone(s) {
		parts = append(parts, "-i", s)
	}
	if s := strings.TrimSpace(in.FQHL); s != "" && !equalNone(s) {
		parts = append(parts, "-I", s)
	}
	parts = append(parts, fmt.Sprintf("%s", mount))
	cmdSet := strings.Join(parts, " ")

	// 若未提供任何限额参数，视为无可更新内容
	if len(parts) == 5 { // 仅基础 5 段: lfs setquota -u 'user' 'mount'
		c.JSON(http.StatusBadRequest, response.Response{Detail: "no quota limits to update"})
		return
	}

	if err := rt.lustreClient.Control(c.Request.Context(), addr, cmdSet); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update user quota: " + err.Error()})
		return
	}

	// 宽限期命令：若提供则更新
	// bg := strings.TrimSpace(in.BQG)
	// ig := strings.TrimSpace(in.FQG)
	// if (bg != "" && !equalNone(bg)) || (ig != "" && !equalNone(ig)) {
	// 	var p2 []string
	// 	p2 = append(p2, "lfs", "setquota", "-t")
	// 	if bg != "" && !equalNone(bg) {
	// 		p2 = append(p2, fmt.Sprintf("--block-grace=%s", bg))
	// 	}
	// 	if ig != "" && !equalNone(ig) {
	// 		p2 = append(p2, fmt.Sprintf("--inode-grace=%s", ig))
	// 	}
	// 	p2 = append(p2, fmt.Sprintf("'%s'", mount))
	// 	cmdGrace := strings.Join(p2, " ")
	// 	if err := rt.lustreClient.Control(c.Request.Context(), addr, cmdGrace); err != nil {
	// 		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update grace: " + err.Error()})
	// 		return
	// 	}
	// }

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

// HandlerUpdateQuota 更新全局配置.
// 执行流程:
//   - 从请求路径中获取 cluster 参数;
//   - 从请求体中获取配额信息 UserQuota（仅使用文件系统与限额字段，忽略 user）；
//   - 构建 lustre 配额配置命令
//     lfs setquota -b <block-softlimit> -B <block-hardlimit> -i <inode-softlimit> -I <inode-hardlimit> <挂载点, 即文件系统>
//     lfs setquota -t --block-grace=xxx --inode-grace=xxx <挂载点, 即文件系统>
//   - 调用 lustreClient.Control 执行上述两个命令.
//
// @Summary 更新 Lustre 全局（默认）配额
// @Description 更新某挂载点的默认配额上限（非用户维度）。参数 cluster 为保留参数。
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param body body UserQuota true "默认配额（仅使用 filesystem 与限额字段，user 忽略）"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quota [put]
func (rt *Router) HandlerUpdateQuota(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析 body
	var in UserQuota
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}
	mount := strings.TrimSpace(in.FileSystem)
	if mount == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "filesystem is required"})
		return
	}

	// 解析转发地址
	addr, err := rt.db.GetLustreServer(cluster)
	if err != nil || strings.TrimSpace(addr) == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 构建 setquota（默认配额）命令：无 -u 选项
	parts := []string{"lfs", "setquota", "-u", "0"}
	base := len(parts)
	if s := strings.TrimSpace(in.BQSL); s != "" && !equalNone(s) {
		parts = append(parts, "-b", s)
	}
	if s := strings.TrimSpace(in.BQHL); s != "" && !equalNone(s) {
		parts = append(parts, "-B", s)
	}
	if s := strings.TrimSpace(in.FQSL); s != "" && !equalNone(s) {
		parts = append(parts, "-i", s)
	}
	if s := strings.TrimSpace(in.FQHL); s != "" && !equalNone(s) {
		parts = append(parts, "-I", s)
	}
	parts = append(parts, fmt.Sprintf("%s", mount))
	if len(parts) == base+1 { // 仅有挂载点，无任何限额参数
		c.JSON(http.StatusBadRequest, response.Response{Detail: "no quota limits to update"})
		return
	}
	cmdSet := strings.Join(parts, " ")
	if err := rt.lustreClient.Control(c.Request.Context(), addr, cmdSet); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update default quota: " + err.Error()})
		return
	}

	// 宽限期（全局）设置
	bg := strings.TrimSpace(in.BQG)
	ig := strings.TrimSpace(in.FQG)
	if (bg != "" && !equalNone(bg)) || (ig != "" && !equalNone(ig)) {
		p2 := []string{"lfs", "setquota", "-t", "-u", "0"}
		if bg != "" && !equalNone(bg) {
			p2 = append(p2, fmt.Sprintf("--block-grace=%s", bg))
		}
		if ig != "" && !equalNone(ig) {
			p2 = append(p2, fmt.Sprintf("--inode-grace=%s", ig))
		}
		p2 = append(p2, fmt.Sprintf("%s", mount))
		cmdGrace := strings.Join(p2, " ")
		if err := rt.lustreClient.Control(c.Request.Context(), addr, cmdGrace); err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update default grace: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

type QuotaApplicationList []QuotaApplication
type QuotaApplication struct {
	ID       int       `json:"id"`
	State    string    `json:"state"`     // 审核状态, 0: 审核中, 1: 完成
	ApplyAt  time.Time `json:"apply_at"`  // 申请日期
	ReviewAt time.Time `json:"review_at"` // 审核日期
	Decision string    `json:"decision"`  // 审核结果
	Apply    UserQuota `json:"apply"`     // 申请内容
	Actual   UserQuota `json:"actual"`    // 当前实际配额
}

var ApplicationStateMaps = map[int]string{
	dbpg.APPLICATION_STATE_PASSED:           "通过",
	dbpg.APPLICATION_STATE_REJECTED:         "拒绝",
	dbpg.APPLICATION_STATE_REVIEWING:        "审核中",
	dbpg.APPLICATION_STATE_PASSED_UNSUCCESS: "通过, 但配额配置失败",
}

// HandlerGetQuotaApps 获取配额申请列表
// API: GET /api/v1/:cluster/lustre/quota/applications?aid=xxx&paging=xxx&page=xxx&page_size=xxx
// 执行流程:
//   - 解析参数, cluster 表示集群名称, aid 表示申请所属用户, 当 aid = -1 时表示获取所有申请.
//     paging 表示是否启动分页, 当 paging = true 时, page, pageSize 有效; 当 paging = false 时表示不分页.
//   - 调用 db.GetQuotaApplications 获取数据. 返回的数据用于填充 QuotaApplicationList 除 Acutal 字段外.
//   - 获取 Actual 所需数据. 调用 lustre client  GetDefaultQuota 获取默认数据，用于填充用户未设置的数据, 调用 lustre client GetUserQuota 获取用户数据.
//   - 根据返回的数据构造QuotaApplicationList, 其中 Application.content 对应解析为 UserQuota.
//
// @Summary 获取配额申请列表
// @Description 获取配额申请列表，支持按申请人筛选与分页；Actual 字段依据 KD 查询当前实际配额并以默认值兜底
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param name query string false "申请人名称, 为空则返回所有申请"
// @Param paging query bool false "是否分页" default(true)
// @Param page query int false "页码，从1开始" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Success 200 {object} response.Response{results=QuotaApplicationList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quota/applications [get]
func (rt *Router) HandlerGetQuotaApps(c *gin.Context) {
	// 1) 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 2) 解析分页参数
	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	pq.SetDefaults(1, 20, 100)

	// 3) 解析 aid，缺省或 -1 表示全部
	applier := strings.TrimSpace(c.Query("name"))

	// 4) 查询申请数据（含总数）
	apps, total, err := rt.db.GetApplications(c.Request.Context(), dbpg.APPLICATION_CLASS_QUOTA, applier, pq.Paging, pq.Page, pq.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch quota applications: " + err.Error()})
		return
	}

	// 5) 解析内容并补充 Actual（需要 slurmrestd 转发 KD）
	addr, err := rt.db.GetLustreServer(cluster)
	if err != nil || strings.TrimSpace(addr) == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	out := make(QuotaApplicationList, 0, len(apps))
	for _, a := range apps {
		qa := QuotaApplication{
			ID:       a.ID,
			State:    ApplicationStateMaps[a.State],
			ApplyAt:  time.Time(a.ApplyAt),
			ReviewAt: time.Time(a.ReviewAt),
			Decision: a.Decision,
		}
		// parse content -> Apply
		if strings.TrimSpace(a.Content) != "" {
			var uq UserQuota
			if err := json.Unmarshal([]byte(a.Content), &uq); err == nil {
				qa.Apply = uq
			}
		}

		// Build Actual based on current KD query with default fallback
		if strings.TrimSpace(qa.Apply.User) != "" && strings.TrimSpace(qa.Apply.FileSystem) != "" {
			dq, err := rt.lustreClient.GetDefaultQuota(c.Request.Context(), addr, qa.Apply.FileSystem)
			if err != nil {
				dq = map[string]string{}
			}
			uq, err := rt.lustreClient.GetUserQuota(c.Request.Context(), addr, qa.Apply.User, qa.Apply.FileSystem)
			if err != nil {
				uq = map[string]string{}
			}
			qa.Actual = UserQuota{
				User:       qa.Apply.User,
				FileSystem: qa.Apply.FileSystem,
				BQSL:       pickWithDefault(uq, dq, LUSTRE_BLOCK_QUOTA_SOFT_LIMIT),
				BQHL:       pickWithDefault(uq, dq, LUSTRE_BLOCK_QUOTA_HARD_LIMIT),
				BQG:        pickWithDefault(uq, dq, LUSTRE_BLOCK_QUOTA_GRACE),
				FQSL:       pickWithDefault(uq, dq, LUSTRE_FILE_QUOTA_SOFT_LIMIT),
				FQHL:       pickWithDefault(uq, dq, LUSTRE_FILE_QUOTA_HARD_LIMIT),
				FQG:        pickWithDefault(uq, dq, LUSTRE_FILE_QUOTA_GRACE),
			}
		}

		out = append(out, qa)
	}

	// 6) 返回（带分页链接）
	if pq.Paging {
		prev, next := response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
		c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: out})
		return
	}
	c.JSON(http.StatusOK, response.Response{Count: total, Results: out})
}

// HandlerGetQuotaAppDecision 获取申请的审核结果.
// API: /api/v1/:cluster/lustre/quota/application/:id/decision
// 执行流程:
//   - 解析参数
//   - 调用 postgres client GetApplicaitionDicision 获取数据
//   - 构造返回响应
//
// @Summary 获取某个配额申请的审核结果
// @Description 获取某个配额申请的审核结果. 参数 cluster 为保留参数, 当前整体设计就不支持多集群.
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param id path int true "申请ID"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quota/application/{id}/decision [get]
func (rt *Router) HandlerGetQuotaAppDecision(c *gin.Context) {
	// 解析 cluster（保持路径规范一致）
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析申请 ID
	sid := strings.TrimSpace(c.Param("id"))
	if sid == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing application id in path"})
		return
	}
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid application id: must be integer"})
		return
	}

	// 调用数据库获取审核结果
	decision, err := rt.db.GetApplicationDecision(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch application decision: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: decision})
}

// HandlerPostQuotaApplication
// API: POST /api/v1/:cluster/lustre/quota/application/:aid
// 执行流程:
//   - 解析参数, aid 表示 applyid.
//   - 解析请求体, 结构为 UserQuota.
//   - 构造 postgres.Application 并调用 AddApplication 函数执行创建申请.
//     构造 postgres.Application 时, applyid = aid, content=json.Marshal(userquota), state = APPLICATION_STATE_REVIEWING
//   - 返回响应.
//
// @Description 参数 cluster 为保留参数, 当前整体设计就不支持多集群.
// @Summary 提交配额申请
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param body body UserQuota true "申请内容（用户配额信息）"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quota/application [post]
func (rt *Router) HandlerCreateQuotaApplication(c *gin.Context) {
	// 解析 cluster（保留参数）
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析申请内容（用户配额）
	var in UserQuota
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}
	b, err := json.Marshal(in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to marshal application content: " + err.Error()})
		return
	}

	// 构造申请并入库
	app := dbpg.Application{
		Applier: in.User,
		Class:   dbpg.APPLICATION_CLASS_QUOTA,
		State:   dbpg.APPLICATION_STATE_REVIEWING,
		Content: string(b),
	}
	if err := rt.db.AddApplication(c.Request.Context(), app); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to add application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

// HandlerUpdateQuotaApplication 更新申请
// API: PUT /api/v1/:cluster/lustre/quota/application/:id
// 执行流程:
//   - 解析参数
//   - 解析请求体(userquota)
//   - 调用 postgres client UpdateQuotaApplication 函数执行更新操作.
//   - 返回响应.
//
// @Summary 更新配额申请
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param id path int true "申请ID"
// @Param body body UserQuota true "申请内容（用户配额信息）"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quota/application/{id} [put]
func (rt *Router) HandlerUpdateQuotaApplication(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析申请 ID
	sid := strings.TrimSpace(c.Param("id"))
	if sid == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing application id in path"})
		return
	}
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid application id: must be integer"})
		return
	}

	// 解析请求体
	var in UserQuota
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}
	b, err := json.Marshal(in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to marshal application content: " + err.Error()})
		return
	}

	// 更新申请
	if err := rt.db.UpdateApplication(c.Request.Context(), id, string(b)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

// HandleDelQuotaApplication
// DELETE /api/v1/:cluster/lustre/quota/application/:id
// 执行流程:
//   - 解析参数
//   - 调用 postgres client DelApplication 函数
//   - 返回响应
//
// @Summary 删除配额申请
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param id path int true "申请ID"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quota/application/{id} [delete]
func (rt *Router) HandleDelQuotaApplication(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析申请 ID
	sid := strings.TrimSpace(c.Param("id"))
	if sid == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing application id in path"})
		return
	}
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid application id: must be integer"})
		return
	}

	if err := rt.db.DelApplication(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to delete application: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

type Review struct {
	Approve   bool   `json:"approve"`  // 是否通过审核
	Descision string `json:"decision"` // 审核意见
	UserQuota
}

// HandlerPostQuotaAppReview 审核申请.
// API: POST /api/v1/:cluster/lustre/quota/application/:id/review
// Request Body: Review
// 执行流程:
//   - 解析参数, id 为 pg.Applications.ID
//   - 若 review.Approve == true, 则:
//     调用 lustre client Control 执行配额配置. 若执行失败则返回失败信息.
//     lfs setquota -u <user> -b <block-softlimit> -B <block-hardlimit> -i <inode-softlimit> -I <inode-hardlimit> <挂载点, 即文件系统>
//     lfs setquota -t --block-grace=xxx --inode-grace=xxx <挂载点, 即文件系统>
//   - db.DoQuotaReview 更新数据库中的申请数据. 若 approve == true, state = APPLICATION_STATE_PASSED. approve==false, state = APPLICATION_STATE_REJECTED
//
// @Summary 审核配额申请
// @Description 根据审核结果执行实际配额配置（通过时）并更新申请的审核状态与内容
// @Tags 资源管理, 存储管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param id path int true "申请ID"
// @Param body body Review true "审核结果与配额内容"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/lustre/quota/application/{id}/review [post]
func (rt *Router) HandlerPostQuotaAppReview(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 解析申请 ID
	sid := strings.TrimSpace(c.Param("id"))
	if sid == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing application id in path"})
		return
	}
	id, err := strconv.Atoi(sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid application id: must be integer"})
		return
	}

	// 解析审核请求体
	var in Review
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}

	// 当通过时，执行实际的配额配置
	if in.Approve {
		user := strings.TrimSpace(in.User)
		mount := strings.TrimSpace(in.FileSystem)
		if user == "" || mount == "" {
			c.JSON(http.StatusBadRequest, response.Response{Detail: "user and filesystem are required when approve=true"})
			return
		}

		addr, err := rt.db.GetLustreServer(cluster)
		if err != nil || strings.TrimSpace(addr) == "" {
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
			}
			return
		}

		// 构建 setquota 命令
		var parts []string
		parts = append(parts, "lfs", "setquota", "-u", fmt.Sprintf("'%s'", user))
		if s := strings.TrimSpace(in.BQSL); s != "" && !equalNone(s) {
			parts = append(parts, "-b", s)
		}
		if s := strings.TrimSpace(in.BQHL); s != "" && !equalNone(s) {
			parts = append(parts, "-B", s)
		}
		if s := strings.TrimSpace(in.FQSL); s != "" && !equalNone(s) {
			parts = append(parts, "-i", s)
		}
		if s := strings.TrimSpace(in.FQHL); s != "" && !equalNone(s) {
			parts = append(parts, "-I", s)
		}
		parts = append(parts, fmt.Sprintf("'%s'", mount))
		cmdSet := strings.Join(parts, " ")

		if len(parts) > 5 { // 有更新项才执行（基础段为5）
			if err := rt.lustreClient.Control(c.Request.Context(), addr, cmdSet); err != nil {
				c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to apply quota: " + err.Error()})
				return
			}
		}

		// 宽限期
		bg := strings.TrimSpace(in.BQG)
		ig := strings.TrimSpace(in.FQG)
		if (bg != "" && !equalNone(bg)) || (ig != "" && !equalNone(ig)) {
			var p2 []string
			p2 = append(p2, "lfs", "setquota", "-t")
			if bg != "" && !equalNone(bg) {
				p2 = append(p2, fmt.Sprintf("--block-grace=%s", bg))
			}
			if ig != "" && !equalNone(ig) {
				p2 = append(p2, fmt.Sprintf("--inode-grace=%s", ig))
			}
			p2 = append(p2, fmt.Sprintf("'%s'", mount))
			cmdGrace := strings.Join(p2, " ")
			if err := rt.lustreClient.Control(c.Request.Context(), addr, cmdGrace); err != nil {
				c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to apply grace: " + err.Error()})
				return
			}
		}
	}

	// 更新审核结果
	decision := dbpg.APPLICATION_STATE_PASSED
	if !in.Approve {
		decision = dbpg.APPLICATION_STATE_REJECTED
	}
	// 仅存储申请的配额内容（不含 Approve 字段）
	contentBytes, err := json.Marshal(in.UserQuota)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to marshal review content: " + err.Error()})
		return
	}
	if err := rt.db.DoReview(c.Request.Context(), id, decision, in.Descision, string(contentBytes)); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update review: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}
