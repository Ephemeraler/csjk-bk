package slurmrest

import (
	"context"
	"csjk-bk/internal/pkg/client/slurmrest/model"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
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
func (sc *Client) GetNodes(ctx context.Context, addr, node string, partitions []string, state []string) (model.Nodes, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	base := fmt.Sprintf("http://%s/api/v1/slurm/nodes", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	if node != "" {
		q.Set("node", node)
	}
	for _, p := range partitions {
		if p != "" {
			q.Add("partition", p)
		}
	}
	for _, s := range state {
		if s != "" {
			q.Add("state", s)
		}
	}
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
func (sc *Client) GetSchedulingJobs(ctx context.Context, addr string, page, pageSize int64) (model.Jobs, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/scheduling/jobs?page=%d&page_page_size=%d", addr, page, pageSize)
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
		Count    int64      `json:"count"`
		Previous url.URL    `json:"previous" swaggertype:"string"`
		Next     url.URL    `json:"next" swaggertype:"string"`
		Results  model.Jobs `json:"results"`
		Detail   string     `json:"detail"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

// GetSchedulingJobs 获取账户中作业信息
func (sc *Client) GetAccountingJobs(ctx context.Context, addr string, page, pageSize int64) (model.Jobs, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/accounting/jobs?page=%d&page_page_size=%d", addr, page, pageSize)
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
		Count    int64      `json:"count"`
		Previous url.URL    `json:"previous" swaggertype:"string"`
		Next     url.URL    `json:"next" swaggertype:"string"`
		Results  model.Jobs `json:"results"`
		Detail   string     `json:"detail"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, 0, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, data.Count, nil
}

func (sc *Client) GetPartitions(ctx context.Context, addr string, partitions []string) ([]map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, sc.timeout)
	defer cancel()

	urlStr := fmt.Sprintf("http://%s/api/v1/slurm/scheduling/partitions", addr)
	if len(partitions) != 0 {
		if u, err := url.Parse(urlStr); err == nil {
			q := u.Query()
			for _, p := range partitions {
				if p == "" {
					continue
				}
				q.Add("partition", p)
			}
			u.RawQuery = q.Encode()
			urlStr = u.String()
		}
	}
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
		Count    int64               `json:"count"`
		Previous url.URL             `json:"previous" swaggertype:"string"`
		Next     url.URL             `json:"next" swaggertype:"string"`
		Results  []map[string]string `json:"results"`
		Detail   string              `json:"detail"`
	}{}

	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		sc.logger.Error("unable to decode slurmrestd response", "err", err.Error(), "url", urlStr)
		return nil, fmt.Errorf("unable to decode slurmrestd response: %w", err)
	}

	return data.Results, nil
}

// GetQos 获取某个 QoS 详情.
func (c *Client) GetQos(ctx context.Context, addr string, id int) (model.QoS, error) {
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

// GetQoSAll 分页获取全部 QoS 信息.
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
