package ldap

import (
	"csjk-bk/internal/pkg/common/paging"
	"csjk-bk/internal/pkg/response"
	"fmt"
	"net/http"
	"strconv"
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

// @Summary 获取某集群 LDAP 用户列表
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param paging query bool false "是否分页" default(true)
// @Param page query int false "页码，从1开始" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Success 200 {object} response.Response{results=[]User}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
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

type AddUser struct {
	Name            string   // 对应 ldap.Uid 必须参数
	CN              string   // 对应 ldap.cn
	SN              string   // 对应 ldap.sn 必须参数
	Passwd          string   // 对应 ldap.userPassword
	Group           int      // 对应 ldap.GidNumber
	AdditionalGroup []string // 对应 ldap.memberuid
	HomeDir         string   // 对应 ldap.homeDirectory
	LoginShell      string   // 对应 ldap.loginShell
	Mobile          string   // 对应 ldap.Mobile
	OU              string   // 对应 ldap.ou
}

// @Summary 在某集群 ldap 中创建用户
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/ldap/user [post]
func (rt *Router) HandlerPostUser(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
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

	// 接收参数
	var in AddUser
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}

	// 基础校验（Uid 已为 int 类型）
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Passwd) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "uid, name and passwd are required"})
		return
	}

	// 将 user 转换成 map[string]string（除 AdditionalGroup 外）
	// 对应的 key 根据注释写成 ldap 的字段
	payload := map[string]string{
		"uid":          in.Name,
		"userPassword": in.Passwd,
		"gidNumber":    fmt.Sprint(in.Group),
	}
	if strings.TrimSpace(in.HomeDir) != "" {
		payload["homeDirectory"] = in.HomeDir
	}
	if strings.TrimSpace(in.LoginShell) != "" {
		payload["loginShell"] = in.LoginShell
	}
	if strings.TrimSpace(in.Mobile) != "" {
		payload["mobile"] = in.Mobile
	}
	if strings.TrimSpace(in.OU) != "" {
		payload["ou"] = in.OU
	}
	if strings.TrimSpace(in.CN) != "" {
		payload["cn"] = in.CN
	}
	if strings.TrimSpace(in.SN) != "" {
		payload["sn"] = in.SN
	}

	// 创建用户
	if err := rt.slurmrestc.AddUser(c.Request.Context(), addr, payload); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to create ldap user: " + err.Error()})
		return
	}

	// 设置用户的附加组（如果提供）
	if len(in.AdditionalGroup) > 0 {
		if err := rt.slurmrestc.AddMemberForGroups(c.Request.Context(), addr, in.Name, in.AdditionalGroup); err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "user created but failed to set additional groups: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

type UpdateUser struct {
	CN         string   // 对应 ldap.cn
	SN         string   // 对应 ldap.sn 必须参数
	Passwd     string   // 对应 ldap.userPassword
	Group      int      // 对应 ldap.GidNumber
	Additional []string // 对应 key 为 additional
	HomeDir    string   // 对应 ldap.homeDirectory
	LoginShell string   // 对应 ldap.loginShell
	Mobile     string   // 对应 ldap.Mobile
	OU         string   // 对应 ldap.ou
}

// @Summary 在某集群 ldap 中更新用户信息
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param name path string true "用户名"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/ldap/user/:name [put]
func (rt *Router) HandlerPutUser(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
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

	// 获取用户名（路径参数）
	name := c.Param("name")
	if strings.TrimSpace(name) == "" {
		// 兼容 :user
		name = c.Param("user")
	}
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing user name in path"})
		return
	}

	// 解析 body
	var in UpdateUser
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}

	// 组装可更新属性
	attr := map[string]string{}

	if strings.TrimSpace(in.CN) != "" {
		attr["cn"] = in.CN
	}
	if strings.TrimSpace(in.SN) != "" {
		attr["sn"] = in.SN
	}
	if strings.TrimSpace(in.Passwd) != "" {
		attr["userPassword"] = in.Passwd
	}
	if in.Group > 0 {
		attr["gidNumber"] = fmt.Sprint(in.Group)
	}
	if strings.TrimSpace(in.HomeDir) != "" {
		attr["homeDirectory"] = in.HomeDir
	}
	if strings.TrimSpace(in.LoginShell) != "" {
		attr["loginShell"] = in.LoginShell
	}
	if strings.TrimSpace(in.Mobile) != "" {
		attr["mobile"] = in.Mobile
	}
	if strings.TrimSpace(in.OU) != "" {
		attr["ou"] = in.OU
	}
	attr["additional"] = strings.Join(in.Additional, ",")

	// 先更新用户属性
	if len(attr) > 0 {
		if err := rt.slurmrestc.UpdateLdapUser(c.Request.Context(), addr, name, attr); err != nil {
			c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update ldap user: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

// @Summary 在某集群 ldap 中删除用户
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param name path string true "用户名"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/ldap/user/:name [delete]
func (rt *Router) HandlerDeleteUser(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
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

	// 获取要删除的用户名（路由中定义为 :name）
	name := c.Param("name")
	if strings.TrimSpace(name) == "" {
		// 兼容可能的 :user 命名
		name = c.Param("user")
	}
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing user name in path"})
		return
	}

	// 调用 slurmrest 客户端删除用户
	if err := rt.slurmrestc.DelLdapUser(c.Request.Context(), addr, name); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to delete ldap user: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

type GroupList []GroupListItem

type GroupListItem struct {
	GID         int      `json:"gid"`         // ldap.GidNumber
	Name        string   `json:"name"`        // ldap.cn
	Description string   `json:"description"` // ldap.description
	Users       []string `json:"users"`       // ldap.memberuid
}

// @Summary 获取 ldap 用户组列表
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param paging query bool false "是否分页" default(true)
// @Param page query int false "页码，从1开始" default(1)
// @Param page_size query int false "每页条数" default(20)
// @Success 200 {object} response.Response{results=GroupList}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/ldap/group/list [get]
func (rt *Router) HandlerGetGroupList(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
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

	// 解析分页参数
	var pq paging.PagingQuery
	_ = c.ShouldBindQuery(&pq)
	if !pq.Paging {
		pq.Paging = true
	}
	pq.SetDefaults(1, 20, 100)

	// 调用 slurmrest 获取组列表
	items, total, err := rt.slurmrestc.GetLdapGroups(c.Request.Context(), addr, pq.Paging, pq.Page, pq.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to fetch ldap groups: " + err.Error()})
		return
	}

	// 转换为 GroupList
	out := make(GroupList, 0, len(items))
	for _, m := range items {
		gidStr := firstNonEmpty(m["gidNumber"], m["gidnumber"], m["gid"], m["GID"])
		gid := 0
		if s := strings.TrimSpace(gidStr); s != "" {
			if v, err := strconv.Atoi(s); err == nil {
				gid = v
			}
		}
		usersRaw := firstNonEmpty(m["memberUid"], m["memberuid"], m["users"], m["Users"])
		var users []string
		if strings.TrimSpace(usersRaw) != "" {
			parts := strings.Split(usersRaw, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					users = append(users, p)
				}
			}
		}
		out = append(out, GroupListItem{
			GID:         gid,
			Name:        firstNonEmpty(m["cn"], m["name"], m["CN"]),
			Description: firstNonEmpty(m["description"], m["desc"], m["Description"]),
			Users:       users,
		})
	}

	prev, next := response.BuildPageLinks(c.Request.URL, pq.Page, pq.PageSize, total)
	c.JSON(http.StatusOK, response.Response{Count: total, Previous: prev, Next: next, Results: out})
}

type AddGroup struct {
	Name        string   `json:"name"`        // 组名
	Description string   `json:"description"` // 组描述
	Users       []string `json:"users"`       // 用户(附加组)
}

// @Summary 在某集群 ldap 中创建用户组
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/ldap/group [post]
func (rt *Router) HandlerAddGroup(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
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

	// 解析 body
	var in AddGroup
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "group name is required"})
		return
	}

	// 组装 payload
	payload := map[string]string{
		"cn":          in.Name,
		"description": in.Description,
	}
	if len(in.Users) > 0 {
		payload["memberUid"] = strings.Join(in.Users, ",")
	}

	// 调用 slurmrest 新增用户组
	if err := rt.slurmrestc.AddGroup(c.Request.Context(), addr, payload); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to add ldap group: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

type UpdateGroup struct {
	Description string   `json:"description"` // 组描述
	Users       []string `json:"users"`       // 用户(附加组)
}

// @Summary 在某集群 ldap 中更新用户组信息
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param name path string true "用户组名称"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/ldap/group/:name [put]
func (rt *Router) HandlerUpdateGroup(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
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

	// 路径中的组名
	name := c.Param("name")
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing group name in path"})
		return
	}

	// 解析 body
	var in UpdateGroup
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "invalid request body: " + err.Error()})
		return
	}

	// 组装更新属性
	attr := map[string]string{}
	if strings.TrimSpace(in.Description) != "" {
		attr["description"] = in.Description
	}
	if len(in.Users) > 0 {
		attr["memberUid"] = strings.Join(in.Users, ",")
	}

	// 若没有任何属性需要更新，直接返回
	if len(attr) == 0 {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "no attributes to update"})
		return
	}

	if err := rt.slurmrestc.UpdateLdapGroup(c.Request.Context(), addr, name, attr); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to update ldap group: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}

// @Summary 在某集群 ldap 中删除用户组
// @Tags 资源管理, 用户管理
// @Produce json
// @Param cluster path string true "集群名称" example("test")
// @Param name path string true "用户组名称"
// @Success 200 {object} response.Response{results=string}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/v1/:cluster/ldap/group/:name [delete]
func (rt *Router) HandlerDeleteGroup(c *gin.Context) {
	// 解析 cluster
	cluster := c.Param("cluster")
	if strings.TrimSpace(cluster) == "" {
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

	// 解析组名
	name := c.Param("name")
	if strings.TrimSpace(name) == "" {
		c.JSON(http.StatusBadRequest, response.Response{Detail: "missing group name in path"})
		return
	}

	// 删除组
	if err := rt.slurmrestc.DelLdapGroup(c.Request.Context(), addr, name); err != nil {
		c.JSON(http.StatusInternalServerError, response.Response{Detail: "failed to delete ldap group: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.Response{Results: "ok"})
}
