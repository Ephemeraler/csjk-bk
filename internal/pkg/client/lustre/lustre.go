package lustre

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	doer   Doer
	logger *slog.Logger
}

func (c *Client) SetClient(doer Doer, logger *slog.Logger) {
	c.doer = doer
	c.logger = logger
}

// Control KD lustre 控制接口.
// POST http://<addr>/api/lustre/execute_cmd
// body: {"command": <cmd> }
// response: {code:int, message: string}
// code == 200 表示响应成功, message 为空. 当 code != 200 时, 表示响应失败, message 为错误详情.
func (c *Client) Control(ctx context.Context, addr, cmd string) error {
	if c.logger == nil {
		panic("lustre client logger == nil")
	}
	if c == nil || c.doer == nil {
		return fmt.Errorf("nil client or http doer")
	}

	// TODO: 本地模拟测试
	c.logger.Debug("Lustre Control Interface", "cmd", cmd)
	return nil

	urlStr := fmt.Sprintf("http://%s/api/lustre/execute_cmd", addr)
	payload := struct {
		Command string `json:"command"`
	}{Command: cmd}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doer.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	// Decode response body regardless of HTTP status to extract server detail
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("decode response failed: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("unexpected http status: %d, detail: %s", resp.StatusCode, data.Message)
	}

	if data.Code != 200 {
		return fmt.Errorf("lustre control failed: code=%d, msg=%s", data.Code, data.Message)
	}

	return nil
}

// Query KD lustre 命令行查询接口.
// GET http://<addr>/api/lustre/execute_cmd?command=<cmd>
// response: {code: int, message: string, result: map[string]string}
// code == 200 表示响应成功, message 为空. 当 code != 200 时, 表示响应失败, message 为错误详情.
func (c *Client) Query(ctx context.Context, addr, cmd string) (map[string]string, error) {
	if c == nil || c.doer == nil {
		return nil, fmt.Errorf("nil client or http doer")
	}

	base := fmt.Sprintf("http://%s/api/lustre/execute_cmd", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("command", cmd)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Result  map[string]string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("unexpected http status: %d, detail: %s", resp.StatusCode, data.Message)
	}

	if data.Code != 200 {
		return nil, fmt.Errorf("lustre query failed: code=%d, msg=%s", data.Code, data.Message)
	}

	return data.Result, nil
}

// GetMounts 获取挂载点.
// 执行流程:
//   - 构建 cmd, df -t lustre
//   - 请求 GET http://<addr>/api/lustre/execute_cmd?command=<cmd>
//   - 解析响应 response: {code: int, message: string, result: []map[string]string}
//   - 返回result
func (c *Client) GetMounts(ctx context.Context, addr string) ([]map[string]string, error) {
	if c == nil || c.doer == nil {
		return nil, fmt.Errorf("nil client or http doer")
	}

	// command to list lustre mounts
	cmd := "df -t lustre"

	// TODO: 本地模拟测试
	c.logger.Debug("Get lustre mounted points", "cmd", cmd)
	return []map[string]string{
		map[string]string{
			"Mounted": "/mnt/lustre",
		},
		map[string]string{
			"Mounted": "/mnt/lustre2",
		},
	}, nil

	base := fmt.Sprintf("http://%s/api/lustre/execute_cmd", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("command", cmd)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Code    int                 `json:"code"`
		Message string              `json:"message"`
		Result  []map[string]string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("unexpected http status: %d, detail: %s", resp.StatusCode, data.Message)
	}
	if data.Code != 200 {
		return nil, fmt.Errorf("lustre get mounts failed: code=%d, msg=%s", data.Code, data.Message)
	}

	return data.Result, nil
}

// GetDefaultQuota 获取 lustre 默认配额
// 执行流程:
//   - 构建 cmd, lfs quota -U <挂载点>
//   - 请求 GET http://<addr>/api/lustre/execute_cmd?command=<cmd>
//   - 解析响应 response: {code: int, message: string, result: map[string]string}
//   - 返回 result
func (c *Client) GetDefaultQuota(ctx context.Context, addr, mount string) (map[string]string, error) {
	if c == nil || c.doer == nil {
		return nil, fmt.Errorf("nil client or http doer")
	}

	cmd := fmt.Sprintf("lfs quota -U %s", mount)
	// TODO: 本地测试
	c.logger.Debug("get default quota for a lustre system", "cmd", cmd)
	return nil, nil

	base := fmt.Sprintf("http://%s/api/lustre/execute_cmd", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("command", cmd)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Result  map[string]string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("unexpected http status: %d, detail: %s", resp.StatusCode, data.Message)
	}
	if data.Code != 200 {
		return nil, fmt.Errorf("lustre get default quota failed: code=%d, msg=%s", data.Code, data.Message)
	}

	return data.Result, nil
}

// GetUserQuota 获取用户配额.
// 执行流程:
//   - 构建 cmd, lfs quota -u <用户名> <挂载点>
//   - 请求 GET http://<addr>/api/lustre/execute_cmd?command=<cmd>
//   - 解析响应 response: {code: int, message: string, result: map[string]string}
//   - 返回 result
func (c *Client) GetUserQuota(ctx context.Context, addr, user, mount string) (map[string]string, error) {
	if c == nil || c.doer == nil {
		return nil, fmt.Errorf("nil client or http doer")
	}

	cmd := fmt.Sprintf("lfs quota -u %s %s", user, mount)

	//TODO: 本地测试
	c.logger.Debug("get quota of user about a lustre system", "cmd", cmd)
	return nil, nil

	base := fmt.Sprintf("http://%s/api/lustre/execute_cmd", addr)
	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("command", cmd)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Result  map[string]string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("unexpected http status: %d, detail: %s", resp.StatusCode, data.Message)
	}
	if data.Code != 200 {
		return nil, fmt.Errorf("lustre get user quota failed: code=%d, msg=%s", data.Code, data.Message)
	}

	return data.Result, nil
}
