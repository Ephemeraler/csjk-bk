package ldap

import (
	"csjk-bk/internal/pkg/common/paging"
	"csjk-bk/internal/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type User struct {
	Uid              string   `json:"uid"`               // 对应 ldap.uidNumber
	Name             string   `json:"name"`              // 对应 ldap.uid
	Group            string   `json:"group"`             // 对应 ldap.gidNumber
	AdditionalGroups []string `json:"additional_groups"` // 需根据 ldap.uid 单独查询
	HomeDirectory    string   `json:"home_dir"`          // 家目录
	CN               string   `json:"cn"`                // 全名
	Mobile           string   `json:"mobile"`            // 电话
	OU               string   `json:"ou"`                // 部门
}

// @Router /api/v1/:cluster/ldap/user/list [get]
func (rt *Router) HandlerGetUserlist(c *gin.Context) {
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

	// 解析分页参数，默认开启分页
	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	if !pq.Paging {
		pq.Paging = true
	}
	pq.SetDefaults(1, 20, 100)

	// 调用 slurmrest 获取用户列表和总数
	items, total, err := rt.slurmrestc.GetLdapUsers(c.Request.Context(), addr, pq.Paging, pq.Page, pq.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch ldap users: " + err.Error()})
		return
	}

	// 组装返回对象，并补充每个用户的附加组
	out := make([]User, 0, len(items))
	for _, m := range items {
		name := firstNonEmpty(m["uid"], m["name"]) // 用户名
		u := User{
			Uid:           firstNonEmpty(m["uidNumber"], m["uidnumber"], m["uid"], m["UID"]),
			Name:          name,
			Group:         firstNonEmpty(m["gidNumber"], m["gidnumber"], m["group"], m["GID"]),
			HomeDirectory: firstNonEmpty(m["homeDirectory"], m["homedirectory"], m["home_dir"], m["home"]),
			CN:            firstNonEmpty(m["cn"], m["common_name"], m["CN"]),
			Mobile:        firstNonEmpty(m["mobile"], m["phone"], m["Mobile"]),
			OU:            firstNonEmpty(m["ou"], m["department"], m["OU"], m["Department"]),
		}
		if strings.TrimSpace(name) != "" {
			groups, err := rt.slurmrestc.GetAdditionalGroupsOfUser(c.Request.Context(), addr, name)
			if err == nil {
				u.AdditionalGroups = groups
			} else {
				// 若查询失败，不影响主流程，保持空切片
				u.AdditionalGroups = []string{}
			}
		}
		out = append(out, u)
	}

	prev, next := response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
	c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: out})
}

// firstNonEmpty returns the first non-empty string from inputs.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
