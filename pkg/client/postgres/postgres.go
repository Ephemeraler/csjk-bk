package postgres

import (
	"context"
	"csjk-bk/pkg/model/alert"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Conditions struct {
	Status   []string
	Start    time.Time
	End      time.Time
	Labels   map[string][]string
	Page     int
	PageSize int
}

type Client struct {
	pool *pgxpool.Pool
}

// GetAlerts 根据报警筛选条件获取报警信息, 条件默认为并集.
func (c *Client) GetAlerts(ctx context.Context, from, to time.Time, status []string, labels, annotations map[string][]string, page, pageSize int64) (alert.Alerts, error) {
	alerts := make(alert.Alerts, 0)
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("无法获取数据库连接: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT id FROM alert WHERE startsat >= $1 AND startsat <= $2 AND status = ANY($3)", from, to, status)
	if err != nil {
		return nil, fmt.Errorf("查询数据库失败: %w", err)
	}

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	// 筛标签
	// SELECT alertid FROM alertlabel WHERE
	// (label = ? AND key = ANY($))

	return alerts, nil
}

// GetAlertLabels 获取标签名列表.
func (c *Client) GetAlertLabels(ctx context.Context) ([]string, error) {
	labels := make([]string, 0)
	rows, err := c.pool.Query(ctx, "SELECT DISTINCT label FROM alertlabel ORDER BY label")
	if err != nil {
		return labels, fmt.Errorf("查询数据库失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			return labels, fmt.Errorf("扫描数据失败: %w", err)
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		return labels, fmt.Errorf("读取数据失败: %w", err)
	}
	return labels, nil
}

// GetAlertLabelValues 获取指定标签名的标签值列表.
func (c *Client) GetAlertLabelValues(ctx context.Context, label string) ([]string, error) {
	values := make([]string, 0)
	rows, err := c.pool.Query(ctx, "SELECT DISTINCT value FROM alertlabel WHERE label = $1 ORDER BY value", label)
	if err != nil {
		return values, fmt.Errorf("查询数据库失败: %w", err)

	}
	defer rows.Close()

	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return values, fmt.Errorf("扫描数据失败: %w", err)
		}
		values = append(values, v)
	}
	if err := rows.Err(); err != nil {
		return values, fmt.Errorf("读取数据失败: %w", err)
	}
	return values, nil
}
