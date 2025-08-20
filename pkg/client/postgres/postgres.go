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
func (c *Client) GetAlerts(ctx context.Context, cond Conditions) (alert.Alerts, error) {
	alerts := make(alert.Alerts, 0)
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("无法获取数据库连接: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT id FROM alert WHERE startsat >= $1 AND startsat <= $2 AND status = ANY($3)", cond.Start, cond.End, cond.Status)
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
