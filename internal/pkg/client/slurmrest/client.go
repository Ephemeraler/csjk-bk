package slurmrest

import (
	"bytes"
	"context"
	"csjk-bk/internal/pkg/client/slurmrest/model"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Doer 抽象 http.Client 的 Do 方法，便于在测试中用 mock 实现替换。
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// SlurmrestClient 简单的 slurmrestd HTTP 客户端封装。
// 仅保留最小必要字段，侧重可测试性与可注入性。
type Client struct {
	client  Doer
	timeout time.Duration
	logger  *slog.Logger
}

func New(client Doer, timeout time.Duration, logger *slog.Logger) *Client {
	return &Client{
		client:  client,
		timeout: timeout,
		logger:  logger,
	}
}

// GetNodes 获取节点信息
// 支持过滤参数：重复传递 partition/state 以及单个 node
//   - partition: ?partition=p1&partition=p2
//   - state: ?state=idle&state=alloc
//   - node: ?node=cn001
func (sc *Client) GetNodes(ctx context.Context, addr string, partitions []string, paging bool, page, page_size int) (model.Nodes, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/scheduling/node/all", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	if len(partitions) != 0 {
		q.Add("partition", strings.Join(partitions, ","))
	}
	q.Add("paging", fmt.Sprintf("%t", paging))
	q.Add("page", fmt.Sprintf("%d", page))
	q.Add("page_size", fmt.Sprintf("%d", page_size))

	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		sc.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return nil, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		sc.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return nil, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results model.Nodes `json:"results"`
		Detail  string      `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return nil, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// GetSchedulingJobs 获取调度作业中作业信息
func (sc *Client) GetSchedulingJobs(ctx context.Context, addr string, paging bool, page, pageSize int) (model.JobsInScheduling, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/scheduling/job/all?paging=%tpage=%d&page_page_size=%d", addr, paging, page, pageSize)
	sc.logger.Debug(urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		sc.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		sc.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int64                  `json:"count"`
		Results model.JobsInScheduling `json:"results"`
		Detail  string                 `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}
	return data.Results, data.Count, nil
}

func (c *Client) GetJobStepsOfScheduling(ctx context.Context, addr, jobid string) (model.JobsStepsInScheduling, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/scheduling/job/steps?jobid=%s", addr, jobid)
	c.logger.Debug(urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return nil, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return nil, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results model.JobsStepsInScheduling `json:"results"`
		Detail  string                      `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}
	return data.Results, nil
}

// GetSchedulingJobs 获取账户中作业信息
func (sc *Client) GetAccountingJobs(ctx context.Context, addr string, paging bool, page, pageSize int64) (model.Jobs, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/accounting/job/all?paging=%t&page=%d&page_page_size=%d", addr, paging, page, pageSize)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		sc.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		sc.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int64      `json:"count"`
		Results model.Jobs `json:"results"`
		Detail  string     `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

// GetPartitions 获取全部分区, 支持分页
func (sc *Client) GetPartitions(ctx context.Context, addr string, paging bool, page, page_size int) ([]map[string]string, int, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/scheduling/partition/all?paging=%t&page=%d&page_size=%d", addr, paging, page, page_size)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		sc.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		sc.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int                 `json:"count"`
		Results []map[string]string `json:"results"`
		Detail  string              `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

func (sc *Client) GetPartitionByName(ctx context.Context, addr string, name string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/scheduling/partition?name=%s", addr, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		sc.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return nil, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		sc.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return nil, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results map[string]string `json:"results"`
		Detail  string            `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// GetQos 获取某个 QoS 详情.
func (c *Client) GetQos(ctx context.Context, addr string, id uint32) (model.QoS, error) {
	// http://addr/api/v1/slurm/accounting/qos?id=xxx 返回类型为 response.Response, 其中 results 的类型就为 model.QoS
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/accounting/qos", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("id", fmt.Sprint(id))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return model.QoS{}, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return model.QoS{}, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return model.QoS{}, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results model.QoS `json:"results"`
		Detail  string    `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return model.QoS{}, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// GetQoSAll 获取全部 QoS 信息, 支持分页.
func (c *Client) GetQosAll(ctx context.Context, addr string, paging bool, page, pageSize int) (model.QoSes, int, error) {
	// http://addr/api/v1/slurm/accounting/qos/all?paging=xxx&page=xxx&page_size=xxx 返回类型为 response.Response, 其中 results 的类型就为 model.QoSes
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/accounting/qos/all", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("paging", fmt.Sprintf("%t", paging))
	if page > 0 {
		q.Set("page", fmt.Sprint(page))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprint(pageSize))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int         `json:"count"`
		Results model.QoSes `json:"results"`
		Detail  string      `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

// GetAccounts 获取全部账户信息, 支持分页.
func (c *Client) GetAccounts(ctx context.Context, add string, paging bool, page, pageSize int) (model.Accounts, int, error) {
	// http://addr/api/v1/slurm/accounting/account/all?paging=xxx&page=xxx&page_size=xxx
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/accounting/account/all", add)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("paging", fmt.Sprintf("%t", paging))
	if page > 0 {
		q.Set("page", fmt.Sprint(page))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprint(pageSize))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int            `json:"count"`
		Results model.Accounts `json:"results"`
		Detail  string         `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

// GetAccount 根据账户名称获取账户信息.
func (c *Client) GetAccountByName(ctx context.Context, addr, name string) (model.Account, error) {
	// http://addr/api/v1/slurm/accounting/account/:name
	// 响应体 response.Response, results 对应 model.Account
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Ensure account name is safely embedded in path
	safeName := url.PathEscape(name)
	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/accounting/account/%s", addr, safeName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return model.Account{}, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return model.Account{}, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return model.Account{}, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results model.Account `json:"results"`
		Detail  string        `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return model.Account{}, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// GetUserByName
// func (c *Client) GetUserByName(ctx context.Context, addr, name string) (model.User, error) {}

type AccountNode struct {
	Name         string     `json:"name"`         // 当前账号节点
	Organization string     `json:"organization"` // 单位
	Description  string     `json:"description"`  // 描述
	SubAccounts  []string   `json:"sub_accounts"` // 子账号名称
	SubUsers     []UserNode `json:"sub_users"`    // 子用户节点信息
}

type UserNode struct {
	Name              string   `json:"name"`               // 用户名
	AdminLevel        int      `json:"admin_level"`        // 管理级别
	AvailableAccounts []string `json:"available_accounts"` // 可用账号
}

// GetChildNodesOfAccount 获取某账户的子节点信息.
func (c *Client) GetChildNodesOfAccount(ctx context.Context, addr, name string) (AccountNode, error) {
	// http://addr/api/v1/slurm/accounting/account/:name/childnodes
	// 返回类型为 response.Response, results 对应类型为 AccountNode
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	safe := url.PathEscape(name)
	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/accounting/account/%s/childnodes", addr, safe)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return AccountNode{}, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return AccountNode{}, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return AccountNode{}, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results AccountNode `json:"results"`
		Detail  string      `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return AccountNode{}, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

type AssociationNode struct {
	Name        string   `json:"name"`         // 账户名称
	Partition   string   `json:"partition"`    // 默认分区
	SubAccounts []string `json:"sub_accounts"` // 子账户
	SubUsers    []AssociationUserNode
}

type AssociationUserNode struct {
	Name       string   `json:"name"` // 用户名
	Partitions []string // 关联分区名称
}

func (c *Client) GetAssociationChildNodesOfAccount(ctx context.Context, addr, name string) (AssociationNode, error) {
	// http://addr/api/v1/slurm/accounting/association/:account/childnodes
	// 返回类型为 response.Response, results 对应类型为 AssociationNode
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	safe := url.PathEscape(name)
	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/accounting/association/%s/childnodes", addr, safe)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return AssociationNode{}, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return AssociationNode{}, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return AssociationNode{}, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results AssociationNode `json:"results"`
		Detail  string          `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return AssociationNode{}, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// GetAssociationDetail 获取某个关联关系, 标识<account, user, partition>
func (c *Client) GetAssociationDetail(ctx context.Context, addr, acct, user, partition string) (model.AssociationItem, error) {
	// http://addr/api/v1/slurm/accounting/association/detail?account=xxx&user=xxx&partition=xxx
	// 返回类型为 response.Response, results 对应类型为 model.AssociationItem
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/accounting/association/detail", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	if acct != "" {
		q.Set("account", acct)
	}
	if user != "" {
		q.Set("user", user)
	}
	if partition != "" {
		q.Set("partition", partition)
	}
	u.RawQuery = q.Encode()

	urlStr := u.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return model.AssociationItem{}, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return model.AssociationItem{}, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return model.AssociationItem{}, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results model.AssociationItem `json:"results"`
		Detail  string                `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return model.AssociationItem{}, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

func (c *Client) GetJobsFromAccounting(ctx context.Context, addr string, paging bool, page, page_size int) (model.Jobs, int, error) {
	// http://addr/api/v1/slurm/accounting/job/all?paging=xxx&page=xxx&page_size=xxx
	// 返回类型为 response.Response, results 对应类型为 model.Jobs
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/accounting/job/all", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("paging", fmt.Sprintf("%t", paging))
	if page > 0 {
		q.Set("page", fmt.Sprint(page))
	}
	if page_size > 0 {
		q.Set("page_size", fmt.Sprint(page_size))
	}
	u.RawQuery = q.Encode()

	urlStr := u.String()
	fmt.Println(urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int        `json:"count"`
		Results model.Jobs `json:"results"`
		Detail  string     `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

func (c *Client) GetJobFromAccounting(ctx context.Context, addr string, jobid uint32) (model.Job, error) {
	// http://addr/api/v1/slurm/accouting/job?jobid=xxx
	// 返回类型为 response.Response, results 对应类型为 model.Job
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/accounting/job", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("jobid", fmt.Sprint(jobid))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return model.Job{}, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return model.Job{}, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return model.Job{}, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results model.Job `json:"results"`
		Detail  string    `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return model.Job{}, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

func (c *Client) GetStepsOfJobFromAccounting(ctx context.Context, addr string, jobid uint32) (model.Steps, error) {
	// http://addr/api/v1/slurm/accounting/job/steps?jobid=xxx
	// 返回类型为 response.Response, results 对应类型为 model.Steps
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/accounting/job/steps", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("jobid", fmt.Sprint(jobid))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return nil, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return nil, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results model.Steps `json:"results"`
		Detail  string      `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return nil, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// GetLdapUsers 返回 ldap 用户及用户总数, 支持分页.
func (c *Client) GetLdapUsers(ctx context.Context, addr string, paging bool, page, pageSize int) ([]map[string]string, int, error) {
	// http://addr/api/v1/ldap/users?paging=xxx&page=xxx&page_size=xxx
	// 返回类型为 response.Response, results 对应类型为 []map[string]string
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/ldap/users", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("paging", fmt.Sprintf("%t", paging))
	if page > 0 {
		q.Set("page", fmt.Sprint(page))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprint(pageSize))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int                 `json:"count"`
		Results []map[string]string `json:"results"`
		Detail  string              `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

// GetAdditionalGroupsOfUser 获取用户的附加组名称.
func (c *Client) GetAdditionalGroupsOfUser(ctx context.Context, addr string, name string) ([]string, error) {
	// http://addr/api/v1/ldap/user/:name/groups
	// 返回类型为 response.Response, results 对应类型为 []string
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	safe := url.PathEscape(name)
	urlStr := fmt.Sprintf("http://%s/api/v1/ldap/user/%s/groups", addr, safe)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return nil, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return nil, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Results []string `json:"results"`
		Detail  string   `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// AddUser 创建 Ldap 用户.
func (c *Client) AddUser(ctx context.Context, addr string, user map[string]string) error {
	// Post http://<addr>/api/v1/ldap/user 请求体直接使用 user(map[string]string)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/ldap/user", addr)

	payload, err := json.Marshal(user)
	if err != nil {
		c.logger.Error("unable to marshal ldap user payload", "err", err.Error())
		return fmt.Errorf("unable to marshal ldap user payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(payload))
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	data := struct {
		Results model.AssociationItem `json:"results"`
		Detail  string                `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return fmt.Errorf("unable to decode response: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr, "detail", data.Detail)
		return fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) UpdateLdapUser(ctx context.Context, addr, user string, attr map[string]string) error {
    // PUT http:/<addr>/api/v1/ldap/user/:user
    // body = attr
    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()

    safe := url.PathEscape(user)
    urlStr := fmt.Sprintf("http://%s/api/v1/ldap/user/%s", addr, safe)

    payload, err := json.Marshal(attr)
    if err != nil {
        c.logger.Error("unable to marshal ldap user update payload", "err", err.Error())
        return fmt.Errorf("unable to marshal ldap user update payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, bytes.NewReader(payload))
    if err != nil {
        c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
        return fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
    }
    defer resp.Body.Close()

    data := struct {
        Detail string `json:"detail"`
    }{}
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
        return fmt.Errorf("unable to decode response: %w", err)
    }

    if resp.StatusCode/100 != 2 {
        c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr, "detail", data.Detail)
        return fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
    }

    return nil
}

func (c *Client) AddMemberForGroups(ctx context.Context, addr string, user string, groups []string) error {
	// http://<addr>/api/v1/ldap/user/:user/groups post
	// body : {"groups": ["g1", "g2", ...]}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	safe := url.PathEscape(user)
	urlStr := fmt.Sprintf("http://%s/api/v1/ldap/user/%s/groups", addr, safe)

	payload := struct {
		Groups []string `json:"groups"`
	}{Groups: groups}

	b, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("unable to marshal groups payload", "err", err.Error())
		return fmt.Errorf("unable to marshal groups payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(b))
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) DelLdapUser(ctx context.Context, addr, user string) error {
	// DELETE http://<addr>/api/v1/ldap/user/:user
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	safe := url.PathEscape(user)
	urlStr := fmt.Sprintf("http://%s/api/v1/ldap/user/%s", addr, safe)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) AddGroup(ctx context.Context, addr string, group map[string]string) error {
	// Post http://<addr>/api/v1/ldap/group
	// body = 参数 group
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/ldap/group", addr)

	payload, err := json.Marshal(group)
	if err != nil {
		c.logger.Error("unable to marshal ldap group payload", "err", err.Error())
		return fmt.Errorf("unable to marshal ldap group payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(payload))
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) DelLdapGroup(ctx context.Context, addr, group string) error {
	// DELETE http://<addr>/api/v1/ldap/group/:group
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	safe := url.PathEscape(group)
	urlStr := fmt.Sprintf("http://%s/api/v1/ldap/group/%s", addr, safe)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, urlStr, nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) UpdateLdapGroup(ctx context.Context, addr, group string, attr map[string]string) error {
	// PUT http://<addr>/api/v1/ldap/group/:group
	// body = attr
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	safe := url.PathEscape(group)
	urlStr := fmt.Sprintf("http://%s/api/v1/ldap/group/%s", addr, safe)

	payload, err := json.Marshal(attr)
	if err != nil {
		c.logger.Error("unable to marshal ldap group update payload", "err", err.Error())
		return fmt.Errorf("unable to marshal ldap group update payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, urlStr, bytes.NewReader(payload))
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", urlStr)
		return fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", urlStr)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", urlStr)
		return fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetLdapGroups(ctx context.Context, addr string, paging bool, page, pageSize int) ([]map[string]string, int, error) {
	// GET http://<addr>/api/v1/ldap/groups?paging=xxx&page=xxx&page_size=xxx
	// 返回类型为 response.Response, results 对应类型为 []map[string]string
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/ldap/groups", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("paging", fmt.Sprintf("%t", paging))
	if page > 0 {
		q.Set("page", fmt.Sprint(page))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprint(pageSize))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		c.logger.Error("unable to create request for slurmrestd", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to create rrequest for slurmrestd: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to do request for slurmrestd", "err", err.Error(), "url", u.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		c.logger.Error("unexcepted status code", "code", resp.StatusCode, "url", u.String())
		return nil, 0, fmt.Errorf("unexcepted status code: %d", resp.StatusCode)
	}

	data := struct {
		Count   int                 `json:"count"`
		Results []map[string]string `json:"results"`
		Detail  string              `json:"detail"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", u.String())
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}
