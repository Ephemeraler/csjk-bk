package slurm

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"csjk-bk/internal/pkg/common/paging"
	"csjk-bk/internal/pkg/common/slurm"
	"csjk-bk/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type Overview struct {
	Nodes       int64 `json:"nodes"`        // 节点总数
	CPUs        int64 `json:"cpus"`         // cpus 总数
	Cores       int64 `json:"cores"`        // 核心总数
	Mems        int64 `json:"mems"`         // 内存总量(MB)
	RunningJobs int64 `json:"running_jobs"` // 运行作业总数
	TotalJobs   int64 `json:"total_jobs"`   // 作业总数
}

// HandlerGetOverview 获取资源统计信息, 包括集群节点总数, 逻辑CPU总数, 总核数, 内存总量, 运行作业数, 总作业数.
// @Summary 获取资源总览页面资源统计信息
// @Description  获取资源总览页面资源统计信息, 包括集群节点总数, 逻辑CPU总数, 总核数, 内存总量, 运行作业数, 总作业数.
// @Tags 资源管理, 资源总览
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=Overview}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/overview [get]
func (rt *Router) HandlerGetOverview(c *gin.Context) {
	// 获取路径参数 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	// 查询 cluster 对应的服务地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	ctx := c.Request.Context()
	// 获取节点(不分页)
	nodes, err := rt.slurmrestc.GetNodes(ctx, addr, nil, false, 0, 0)
	if err != nil {
		rt.logger.Error("unable to get nodes information", "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "无法获取节点信息数据"})
		return
	}

	_, runningCount, err := rt.slurmrestc.GetSchedulingJobs(ctx, addr, true, 1, 1)
	if err != nil {
		rt.logger.Error("unable to get job information from slurm scheduling", "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "无法获取运行作业数据"})
		return
	}

	_, totalCount, err := rt.slurmrestc.GetAccountingJobs(ctx, addr, true, 1, 1)
	if err != nil {
		rt.logger.Error("unable to get job information from slurm accouting", "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "无法获取账户系统中作业信息"})
		return
	}

	// 计算节点总数、CPU、核心、内存
	var (
		nodeNum int64
		cpus    int64
		cores   int64
		mems    int64
	)
	nodeNum = int64(len(nodes))
	for _, n := range nodes {
		cpus += n.CPUs
		cores += n.Cores
		mems += n.Memory
	}

	ov := Overview{
		Nodes:       nodeNum,
		CPUs:        cpus,
		Cores:       cores,
		Mems:        mems,
		RunningJobs: runningCount,
		TotalJobs:   totalCount,
	}
	c.JSON(http.StatusOK, response.Response{Results: ov})
}

type NodeDetails map[string]NodeDetail // key 为节点名称
type NodeDetail struct {
	Name      string `json:"name"`      // 节点名称
	State     string `json:"state"`     // 节点状态
	CPUs      int64  `json:"cpus"`      // cpus 总数
	Mems      int64  `json:"mems"`      // 内存总量(MB)
	GPU       string `json:"gpu"`       // gpu 详情
	Partition string `json:"partition"` // 分区名称
}

type PartitionsList []PartitionListItem

type PartitionListItem struct {
	Name     string `json:"name"`     // 分区名称
	IsDef    string `json:"is_def"`   // 是否为默认分区
	State    string `json:"state"`    // 分区状态
	MaxTime  string `json:"max_time"` // 最长作业时间限制
	NodeNum  string `json:"node_num"` // 节点数
	Nodelist string `json:"nodelist"` // 节点列表
}

// HandlerGetPartitionList 获取所有分区概述
// @Summary 获取某集群所有分区概述
// @Tags 资源管理, 分区管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param paging query bool false "是否开启分页" default(true)
// @Param page query int false "页号(从1开始)" example("1") default(1) minimum(1)
// @Param page_size query int false "每页数量" example("20") default(20) minimum(1) maximum(100)
// @Success 200 {object} response.Response{results=PartitionsList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/partition/list [get]
func (rt *Router) HandlerGetPartitionList(c *gin.Context) {
	list := make(PartitionsList, 0)
	// 获取路径参数 cluster
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

	// 获取 Partition List
	partitions, total, err := rt.slurmrestc.GetPartitions(c.Request.Context(), addr, pq.Paging, pq.Page, pq.PageSize)
	if err != nil {
		rt.logger.Error("unable to get all partitions information", "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "无法获取 Partition 元数据"})
		return
	}

	for _, partition := range partitions {
		item := PartitionListItem{
			Name:     partition["PartitionName"],
			IsDef:    partition["Default"],
			State:    partition["State"],
			MaxTime:  partition["MaxTime"],
			NodeNum:  partition["TotalNodes"],
			Nodelist: partition["Nodes"],
		}
		list = append(list, item)
	}

	var prevURL, nextURL url.URL
	if pq.Paging {
		prevURL, nextURL = response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
	}

	c.JSON(http.StatusOK, response.Response{Count: total, Previous: prevURL, Next: nextURL, Results: list})
}

type PartitionDetail struct {
	Name            string `json:"name"`               // 分区名称, 对应 PartitionName
	IsDef           string `json:"is_def"`             // 是否为默认分区, 对应 Default
	State           string `json:"state"`              // 状态, 对应 State
	Hidden          string `json:"is_hidden"`          // 是否为隐藏分区, 对应 Hidden
	Alternate       string `json:"alertnate"`          // 备注(选)分区, 对应 Alternate
	IsResv          string `json:"is_resv"`            // 是否需要预约, 对应 ReqResv
	CPUNum          string `json:"cpu_num"`            // CPU数, 从节点信息中统计
	NodeNum         string `json:"node_num"`           // 节点数, 从节点信息统计
	AllowAllocNodes string `json:"allow_alloc_nodes"`  // 允许分配的节点数, 对应 AllowAllocNodes(更优先) 或 Nodes
	MaxCPUsPerNode  string `json:"max_cpus_per_nodes"` // 每节点最多 CPU 数, 对应 MaxCPUsPerNode
	DefMemPerCPU    string `json:"def_mem_per_cpu"`    // 默认每 cpu 分配的内存数, 对应 DefMemPerCPU
	DefMemPerNode   string `json:"def_mem_per_node"`   // 默认每节点分配的内存数, 对应 DefMemPerNode
	MaxMemPerCPU    string `json:"max_mem_per_cpu"`    // 最大每 cpu 分配的内存数, 对应 MaxMemPerCPU
	MaxMemPerNode   string `json:"max_mem_per_node"`   // 最大每节点分配内存数, 对应 MaxMemPerNode
	CPUBindParam    string `json:"cpu_bind_param"`     // CPU 绑定参数
	SuspendTimeout  string `json:"suspend_timeout"`    // 节点挂起超时, 对应 SuspendTimeout
	SuspendTime     string `json:"suspend_time"`       // 节点挂起前空闲时间, 对应 SuspendTime
	Nodelist        string `json:"nodelist"`           // 节点列表, 对应 Nodes
	AllowGroups     string `json:"allow_groups"`       // 允许的用户组, 对应 AllowGroups
	AllowAccounts   string `json:"allow_accounts"`     // 允许的账户, 对应 AllowAccounts
	DenyAccounts    string `json:"deny_accounts"`      // 禁止的账户, 对应 DenyAccounts
	RootOnly        string `json:"root_only"`          // 是否仅允许 root 提交作业, 对应 RootOnly
	ExclusiveUser   string `json:"exclusive_user"`     // 是否用户独占, 对应 ExclusiveUser
	// 缺失, 是否禁止 root 作业
	OverSubscribe string `json:"over_subscribe"` // 是否按负载分配节点
}

// HandlerGetPartitionDetail 获取某分区详情
// 实现流程:
//   - 获取路径参数 cluster, 并通过 db 获取 solid 服务地址
//   - 获取集群详情及集群节点详情
//
// @Summary 获取某集群中某个分区的详细内容
// @Tags 资源管理, 分区管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param name path string true "分区名称" example("p1")
// @Success 200 {object} response.Response{results=PartitionDetail}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/partition/{name}/detail [get]
func (rt *Router) HandlerGetPartitionDetail(c *gin.Context) {
	// 参数解析
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}
	name := c.Param("name")
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing required query: partition"})
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

	partition, err := rt.slurmrestc.GetPartitionByName(c.Request.Context(), addr, name)
	if err != nil {
		// TODO 存在分区不存在的情况也会报这个错误.
		rt.logger.Error("unable to get partition informatio", "err", err)
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "无法获取分区详情"})
		return
	}

	pd := PartitionDetail{
		Name:            partition["PartitionName"],
		IsDef:           partition["Default"],
		State:           partition["State"],
		Hidden:          partition["Hidden"],
		Alternate:       partition["Alternate"],
		IsResv:          partition["ReqResv"],
		CPUNum:          partition["TotalCPUs"],
		NodeNum:         partition["TotalNodes"],
		AllowAllocNodes: partition["AllocNodes"],
		MaxCPUsPerNode:  partition["MaxCPUsPerNode"],
		DefMemPerCPU:    partition["DefMemPerCPU"],
		DefMemPerNode:   partition["DefMemPerNode"],
		MaxMemPerCPU:    partition["MaxMemPerCPU"],
		MaxMemPerNode:   partition["MaxMemPerNode"],
		CPUBindParam:    partition["CPUBindParam"],
		SuspendTimeout:  partition["SuspendTimeout"],
		SuspendTime:     partition["SuspendTime"],
		Nodelist:        partition["Nodes"],
		AllowGroups:     partition["AllowGroups"],
		AllowAccounts:   partition["AllowAccounts"],
		DenyAccounts:    partition["DenyAccounts"],
		RootOnly:        partition["RootOnly"],
		ExclusiveUser:   partition["ExclusiveUser"],
		OverSubscribe:   partition["OverSubscribe"],
	}

	c.JSON(http.StatusOK, response.Response{Results: pd})
}

type LdapUsers []LdapUser
type LdapUser struct {
	UID           string `json:"UID"`              // ldap uidNumber
	Name          string `json:"name"`             // 用户名
	PrimaryGrp    string `json:"primary_group"`    // 主组
	AdditionalGrp string `json:"additional_group"` // 附加组
	HomeDir       string `json:"home_dir"`         // HOME 目录
	CommonName    string `json:"common_name"`      // 全名
	Mobile        string `json:"mobile"`           // 电话
	Department    string `json:"Department"`       // 部门
}

// @Summary 获取某集群中 LDAP 所有用户
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=LdapUsers}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/ldap/users [get]
func (rt *Router) HandlerGetLdapUsers(c *gin.Context) {}

// @Summary 在某集群 ldap 中创建用户
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/ldap/user [post]
func (rt *Router) HandlerCreateLdapUser(c *gin.Context) {}

// @Summary 在某集群 ldap 中更新用户信息
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/ldap/user [put]
func (rt *Router) HandlerUpdateLdapUser(c *gin.Context) {}

// @Summary 在某集群 ldap 中删除某用户
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/ldap/user [delete]
func (rt *Router) HandlerDeleteLdapUser(c *gin.Context) {}

// @Summary 获取某集群中 slurm 账户名列表
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=AccountNameList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/accounting/accounts/name [get]
func (rt *Router) HandlerGetAccountingAcctsname(c *gin.Context) {}

type LdapGroups []LdapGroup
type LdapGroup struct {
	GID   string `json:"gid"`   // 用户组 id(linux gid), 对应 ldap gidNumber
	Name  string `json:"name"`  // 用户组名称, 对应 ldap gid
	Users string `json:"users"` // 用户
}

// @Summary 获取某集群中 ldap 用户组列表
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=LdapGroups}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/ldap/groups [get]
func (rt *Router) HandlerGetLdapGroups(c *gin.Context) {}

// @Summary 向某集群中 ldap 用户组添加新用户组
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/ldap/group [post]
func (rt *Router) HandlerPostLdapGroups(c *gin.Context) {}

// @Summary 删除某集群中 ldap 用户组
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/ldap/group [delete]
func (rt *Router) HandlerDeleteLdapGroups(c *gin.Context) {}

type JobListOfScheduling []JobListItemOfScheduling

type JobListItemOfScheduling struct {
	Jobid     string `json:"jobid"`     // 作业 ID
	State     string `json:"state"`     // 状态
	User      string `json:"user"`      // 账户 user(account)
	CPUs      string `json:"cpus"`      // 资源个数
	Nodelist  string `json:"nodelist"`  // 节点列表
	Partition string `json:"partition"` // 分区
	QoS       string `json:"qos"`       // QoS
	Reason    string `json:"reason"`    // 原因
}

// HandlerGetSchedulingJobList 获取某集群调度队列中的作业列表
// @Summary 获取某集群调度列表中作业列表
// @Description 返回当前调度队列中的作业列表，可选分页参数
// @Tags 资源管理, 作业管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param paging query bool false "是否开启分页, 当前仅支持true" default(true)
// @Param page query int false "页号(从1开始)" example("1") default(1) minimum(1)
// @Param page_size query int false "每页数量" example("20") default(20) minimum(1)
// @Success 200 {object} response.Response{results=JobListOfScheduling}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/scheduling/job/list [get]
func (rt *Router) HandlerGetSchedulingJobList(c *gin.Context) {
	list := make(JobListOfScheduling, 0)

	// 解析 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Results: list, Detail: "missing cluster in path"})
		return
	}

	// 解析分页参数
	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	pq.SetDefaults(1, 20, 100)
	pq.Paging = true

	// 获取 slurmrestd 地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Results: list, Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Results: list, Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 查询调度作业
	items, total, err := rt.slurmrestc.GetSchedulingJobs(c.Request.Context(), addr, pq.Paging, pq.Page, pq.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Results: list, Detail: "failed to fetch scheduling jobs: " + err.Error()})
		return
	}

	for _, item := range items {
		// 获取 QoS 名称，失败则回退为 ID
		list = append(list, JobListItemOfScheduling{
			Jobid:     item.Jobid,
			State:     item.State,
			User:      fmt.Sprintf("%d(%s)", item.User, item.Account),
			CPUs:      item.CPUs,
			Nodelist:  item.Nodelist,
			Partition: item.Partition,
			QoS:       item.QoS,
			Reason:    item.Reason,
		})
	}

	var prev, next url.URL
	if pq.Paging {
		prev, next = response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, int(total))
	}

	c.JSON(http.StatusOK, response.Response{Count: int(total), Previous: prev, Next: next, Results: list})
}

// @Summary 获取某集群调度列表中某作业详情
// @Tags 资源管理, 作业管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param jobid path int true "作业号" example("1")
// @Success 200 {object} response.Response{results=model.JobsStepsInScheduling}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/scheduling/job/{jobid}/detail [get]
func (rt *Router) HandlerGetSchedulingJobsDetail(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
		return
	}

	jobid := c.Param("jobid")
	if jobid == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing jobid in path"})
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

	detail, err := rt.slurmrestc.GetJobStepsOfScheduling(c.Request.Context(), addr, jobid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: fmt.Sprintf("%s", err)})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: detail})
}

type JobListOfAccounting []JobListItem
type JobListItem struct {
	JobID     uint32 `json:"jobid"`      // 作业ID
	State     string `json:"state"`      // 状态
	Account   string `json:"account"`    // 账号
	TresAlloc string `json:"tres_alloc"` // 资源个数
	Nodelist  string `json:"nodelist"`   // 节点列表
	Partition string `json:"partition"`  // 分区
	QoS       string `json:"qos"`        // QoS
	Reason    string `json:"reason"`     // 原因
}

// @Summary 获取某集群账户中作业列表
// @Description 返回账户历史作业列表，可选分页参数
// @Tags 资源管理, 作业管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param paging query bool false "是否开启分页, 当前仅支持true" default(true)
// @Param page query int false "页号(从1开始)" example("1") default(1) minimum(1)
// @Param page_size query int false "每页数量" example("20") default(20) minimum(1)
// @Success 200 {object} response.Response{results=JobListOfAccounting}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/accounting/job/list [get]
func (rt *Router) HandlerGetJobListFromAccounting(c *gin.Context) {
	list := make(JobListOfAccounting, 0)

	// 解析 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Results: list, Detail: "missing cluster in path"})
		return
	}

	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	pq.SetDefaults(1, 20, 100)
	pq.Paging = true
	// 获取 slurmrestd 地址
	addr, err := rt.db.GetSlurmrestdAddr(cluster)
	if err != nil || addr == "" {
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Results: list, Detail: "failed to resolve slurmrestd address: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, response.Response{Results: list, Detail: "empty slurmrestd address for cluster"})
		}
		return
	}

	// 查询作业
	items, total, err := rt.slurmrestc.GetJobsFromAccounting(c.Request.Context(), addr, pq.Paging, pq.Page, pq.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Results: list, Detail: "failed to fetch accounting jobs: " + err.Error()})
		return
	}

	for _, item := range items {
		if qosName, err := rt.slurmrestc.GetQos(c.Request.Context(), addr, item.IDQOS); err != nil {
			list = append(list, JobListItem{
				JobID:     item.IDJob,
				State:     slurm.PrintJobStateString(item.State),
				Account:   item.Account,
				TresAlloc: item.TresAlloc,
				Nodelist:  item.Nodelist,
				Partition: item.Partition,
				QoS:       fmt.Sprintf("%d", item.IDQOS),
				Reason:    slurm.PrintJobStateReasonStr(item.StateReasonPrev),
			})
		} else {
			list = append(list, JobListItem{
				JobID:     item.IDJob,
				State:     slurm.PrintJobStateString(item.State),
				Account:   item.Account,
				TresAlloc: item.TresAlloc,
				Nodelist:  item.Nodelist,
				Partition: item.Partition,
				QoS:       qosName.Name,
				Reason:    slurm.PrintJobStateReasonStr(item.StateReasonPrev),
			})
		}
	}

	var prev, next url.URL
	if pq.Paging {
		prev, next = response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
	}

	// 仅返回当前页数据；Count 使用当前条数
	c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: list})
}

type DetailOfJobFromAccounting struct {
	Jobid                    uint32                      `json:"jobid"` // 作业ID
	StepsOfJobFromAccounting []StepOfJobFromAccounting   `json:"steps"`
	Meta                     MetadataOfJobFromAccounting `json:"meta"`
}

type StepOfJobFromAccounting struct {
	Name  string `json:"name"`  // 作业步名称
	State string `json:"state"` // 作业步状态
}

type MetadataOfJobFromAccounting struct {
	User       string `json:"user"`        // 用户名
	Group      string `json:"group"`       // 分组
	Account    string `json:"account"`     // 账号
	Priority   uint32 `json:"priority"`    // 优先级
	JobName    string `json:"job_name"`    // 作业名
	WorkDir    string `json:"work_dir"`    // 工作目录
	JobCmd     string `json:"job_cmd"`     // 作业命令
	Result     string `json:"result"`      // 运行结果 state 和 exitcode 组合
	NodesAlloc uint32 `json:"nodes_alloc"` // 节点数
	Partition  string `json:"partition"`   // 分区
	QoS        string `json:"qos"`         // QoS 名称
}

// HandlerGetAccountingJobDetail 获取某集群账户中某作业详情
// @Summary 获取某集群账户中某作业详情
// @Description 返回作业基础元信息与所有作业步列表
// @Tags 资源管理, 作业管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param jobid path int true "作业号" example("1")
// @Success 200 {object} response.Response{results=DetailOfJobFromAccounting}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/accounting/job/{jobid}/detail [get]
func (rt *Router) HandlerGetAccountingJobDetail(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if cluster == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing cluster in path"})
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

	// 解析 jobid
	jobidStr := c.Param("jobid")
	jobid64, err := strconv.ParseUint(jobidStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid jobid in path"})
		return
	}
	jobid := uint32(jobid64)

	// 查询作业与步骤
	job, err := rt.slurmrestc.GetJobFromAccounting(c.Request.Context(), addr, jobid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch job from accounting: " + err.Error()})
		return
	}
	steps, err := rt.slurmrestc.GetStepsOfJobFromAccounting(c.Request.Context(), addr, jobid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch steps of job from accounting: " + err.Error()})
		return
	}

	// 构造 steps 结果
	stepsOut := make([]StepOfJobFromAccounting, 0, len(steps))
	for _, s := range steps {
		stepsOut = append(stepsOut, StepOfJobFromAccounting{
			Name:  s.StepName,
			State: slurm.PrintJobStateString(s.State),
		})
	}

	// QoS 名称查询，失败则回退为 ID 字符串
	qosName := fmt.Sprintf("%d", job.IDQOS)
	if q, e := rt.slurmrestc.GetQos(c.Request.Context(), addr, job.IDQOS); e == nil && q.Name != "" {
		qosName = q.Name
	}

	// 组合结果
	detail := DetailOfJobFromAccounting{
		Jobid:                    job.IDJob,
		StepsOfJobFromAccounting: stepsOut,
		Meta: MetadataOfJobFromAccounting{
			// TODO 未来使用 ldap 客户端查询uid和gid对应的名称.
			User:       fmt.Sprintf("%d", job.IDUser),
			Group:      fmt.Sprintf("%d", job.IDGroup),
			Account:    job.Account,
			Priority:   job.Priority,
			JobName:    job.JobName,
			WorkDir:    job.WorkDir,
			JobCmd:     strings.TrimSpace(job.SubmitLine),
			Result:     fmt.Sprintf("%s(%d)", slurm.PrintJobStateString(job.State), job.ExitCode),
			NodesAlloc: job.NodesAlloc,
			Partition:  job.Partition,
			QoS:        qosName,
		},
	}

	c.JSON(http.StatusOK, response.Response{Results: detail})
}
