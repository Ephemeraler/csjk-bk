package slurm

import (
	"net/http"

	"csjk-bk/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type AccountNameList []string

// HandlerGetAccountsNameList 获取某集群中所有 Slurm 账户名称列表。
//
// @Summary 获取某集群账户名称列表
// @Description 返回账户名称字符串数组，无分页
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=AccountNameList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/account/name/list [get]
func (rt *Router) HandlerGetAccountsNameList(c *gin.Context) {
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

	// 调用 slurmrest 客户端，获取当前页数据
	items, total, err := rt.slurmrestc.GetAccounts(c.Request.Context(), addr, false, 0, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch accounts list: " + err.Error()})
		return
	}

	// 组装输出列表
	out := make(AccountNameList, 0, len(items))
	for _, it := range items {
		out = append(out, it.Name)
	}
	// 返回带分页信息的响应
	// 注意：当 paging=false 时，prev/next 为空，count 为列表长度
	c.JSON(http.StatusOK, response.Response{Count: total, Results: out})
}

// @Param account path string true "账户节点名称" example("test")
// @Summary 获取某集群中指定账户的子节点信息
// @Description 返回账户节点的子账号名称列表及子用户节点信息
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param account path string true "账户节点名称" example("root")
// @Success 200 {object} response.Response{results=slurmrest.AccountNode}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/{cluster}/slurm/account/{account}/childnodes [get]
func (rt *Router) HandlerGetAccountChildNodes(c *gin.Context) {
	// 解析路径参数
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
	node, err := rt.slurmrestc.GetChildNodesOfAccount(c.Request.Context(), addr, account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch account child nodes: " + err.Error()})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, response.Response{Results: node})
}
