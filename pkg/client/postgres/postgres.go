package postgres

import (
	"context"
	"csjk-bk/pkg/model/alert"
	"fmt"
	"strings"
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

	var sb strings.Builder
	sb.WriteString("SELECT a.id, a.status, a.startsat, a.endsat FROM alert a WHERE 1=1")

	args := make([]interface{}, 0)
	idx := 1

	if !from.IsZero() {
		sb.WriteString(fmt.Sprintf(" AND a.startsat >= $%d", idx))
		args = append(args, from)
		idx++
	}
	if !to.IsZero() {
		sb.WriteString(fmt.Sprintf(" AND a.startsat <= $%d", idx))
		args = append(args, to)
		idx++
	}
	if len(status) > 0 {
		sb.WriteString(fmt.Sprintf(" AND a.status = ANY($%d)", idx))
		args = append(args, status)
		idx++
	}

	for k, vals := range labels {
		if len(vals) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf(" AND EXISTS (SELECT 1 FROM alertlabel al WHERE al.alertid = a.id AND al.label = $%d AND al.value = ANY($%d))", idx, idx+1))
		args = append(args, k, vals)
		idx += 2
	}

	for k, vals := range annotations {
		if len(vals) == 0 {
			continue
		}
		patterns := make([]string, 0, len(vals))
		for _, v := range vals {
			patterns = append(patterns, "%"+v+"%")
		}
		sb.WriteString(fmt.Sprintf(" AND EXISTS (SELECT 1 FROM alertannotation an WHERE an.alertid = a.id AND an.annotation = $%d AND an.value ILIKE ANY($%d))", idx, idx+1))
		args = append(args, k, patterns)
		idx += 2
	}

	sb.WriteString(" ORDER BY a.startsat DESC")
	if pageSize > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT $%d", idx))
		args = append(args, pageSize)
		idx++
		if page > 0 {
			sb.WriteString(fmt.Sprintf(" OFFSET $%d", idx))
			args = append(args, (page-1)*pageSize)
			idx++
		}
	}

	rows, err := conn.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("查询数据库失败: %w", err)
	}
	defer rows.Close()

	idMap := make(map[int64]*alert.Alert)
	ids := make([]int64, 0)
	for rows.Next() {
		a := &alert.Alert{Label: make(map[string]string), Annotation: make(map[string]string)}
		if err := rows.Scan(&a.ID, &a.Status, &a.StartsAt, &a.EndsAt); err != nil {
			return nil, fmt.Errorf("读取数据失败: %w", err)
		}
		alerts = append(alerts, a)
		idMap[a.ID] = a
		ids = append(ids, a.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("读取数据失败: %w", err)
	}

	if len(ids) == 0 {
		return alerts, nil
	}

	lrows, err := conn.Query(ctx, "SELECT alertid, label, value FROM alertlabel WHERE alertid = ANY($1)", ids)
	if err != nil {
		return nil, fmt.Errorf("查询数据库失败: %w", err)
	}
	for lrows.Next() {
		var id int64
		var k, v string
		if err := lrows.Scan(&id, &k, &v); err != nil {
			lrows.Close()
			return nil, fmt.Errorf("读取数据失败: %w", err)
		}
		if al, ok := idMap[id]; ok {
			al.Label[k] = v
		}
	}
	if err := lrows.Err(); err != nil {
		lrows.Close()
		return nil, fmt.Errorf("读取数据失败: %w", err)
	}
	lrows.Close()

	arows, err := conn.Query(ctx, "SELECT alertid, annotation, value FROM alertannotation WHERE alertid = ANY($1)", ids)
	if err != nil {
		return nil, fmt.Errorf("查询数据库失败: %w", err)
	}
	for arows.Next() {
		var id int64
		var k, v string
		if err := arows.Scan(&id, &k, &v); err != nil {
			arows.Close()
			return nil, fmt.Errorf("读取数据失败: %w", err)
		}
		if al, ok := idMap[id]; ok {
			al.Annotation[k] = v
		}
	}
	if err := arows.Err(); err != nil {
		arows.Close()
		return nil, fmt.Errorf("读取数据失败: %w", err)
	}
	arows.Close()

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
