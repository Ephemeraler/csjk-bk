package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool 抽象出连接池接口，便于在单元测试中通过自定义实现进行 mock。
// 该接口与 pgxpool.Pool 的常用子集保持一致，满足多数查询/执行需求。
type Pool interface {
	Ping(ctx context.Context) error
	Close()
	Acquire(ctx context.Context) (c *pgxpool.Conn, err error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Client Postgres 客户端，内部使用连接池。
type Client struct {
	pool Pool
}

// New 根据 DSN 创建 Postgres 客户端，内部使用 pgxpool 连接池。
// 示例 DSN："postgres://user:pass@localhost:5432/dbname?sslmode=disable"
func New(ctx context.Context, dsn string, opts ...Option) (*Client, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	for _, o := range opts {
		o(cfg)
	}
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	// 建立连接并验证连通性
	if err := p.Ping(ctx); err != nil {
		p.Close()
		return nil, err
	}
	return &Client{pool: p}, nil
}

// NewWithPool 允许注入自定义 Pool（用于单元测试的 mock）。
func NewWithPool(p Pool) *Client { return &Client{pool: p} }

// Pool 返回底层连接池接口，便于执行查询或扩展能力。
func (c *Client) Pool() Pool { return c.pool }

// Close 关闭底层连接池。
func (c *Client) Close() {
	if c != nil && c.pool != nil {
		c.pool.Close()
	}
}

// GetSlurmrestAddr 获取 slurmrestd 服务地址
func (c *Client) GetSlurmrestdAddr(cluster string) (string, error) {
	return "192.168.2.35:39999", nil
}

type Alerts []*Alert

type Alert struct {
	ID         int64
	Status     string
	StartsAt   time.Time
	EndsAt     time.Time
	Responder  string
	Operation  string
	Label      map[string]string
	Annotation map[string]string
}

// GetAlerts 查询报警记录，支持按时间范围、状态、标签与注释过滤，并支持分页。
// 新增返回值 total 用于表示在分页前、按过滤条件筛选后的记录总数。
func (c *Client) GetAlerts(ctx context.Context, from, to time.Time, status []string, labels, annotations map[string][]string, page, pageSize int) (Alerts, int, error) {
	alerts := make(Alerts, 0)
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("无法获取数据库连接: %w", err)
	}
	defer conn.Release()

	// 构建通用过滤子句（供 count 与列表查询复用）
	var whereSB strings.Builder
	whereSB.WriteString(" WHERE 1=1")

	args := make([]interface{}, 0)
	idx := 1

	if !from.IsZero() {
		whereSB.WriteString(fmt.Sprintf(" AND a.startsat >= $%d", idx))
		args = append(args, from)
		idx++
	}
	if !to.IsZero() {
		whereSB.WriteString(fmt.Sprintf(" AND a.startsat <= $%d", idx))
		args = append(args, to)
		idx++
	}
	if len(status) > 0 {
		whereSB.WriteString(fmt.Sprintf(" AND a.status = ANY($%d)", idx))
		args = append(args, status)
		idx++
	}

	for k, vals := range labels {
		if len(vals) == 0 {
			continue
		}
		whereSB.WriteString(fmt.Sprintf(" AND EXISTS (SELECT 1 FROM alertlabel al WHERE al.alertid = a.id AND al.label = $%d AND al.value = ANY($%d))", idx, idx+1))
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
		whereSB.WriteString(fmt.Sprintf(" AND EXISTS (SELECT 1 FROM alertannotation an WHERE an.alertid = a.id AND an.annotation = $%d AND an.value ILIKE ANY($%d))", idx, idx+1))
		args = append(args, k, patterns)
		idx += 2
	}

	// 先查询总数（无分页）
	countSQL := "SELECT COUNT(*) FROM alert a" + whereSB.String()
	var total int64
	if err := conn.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计总数失败: %w", err)
	}

	// 再查询列表（带排序与分页）
	var listSB strings.Builder
	listSB.WriteString("SELECT a.id, a.status, a.startsat, a.endsat FROM alert a")
	listSB.WriteString(whereSB.String())
	listSB.WriteString(" ORDER BY a.startsat DESC")

	listArgs := make([]interface{}, len(args))
	copy(listArgs, args)
	lidx := idx
	if pageSize > 0 {
		listSB.WriteString(fmt.Sprintf(" LIMIT $%d", lidx))
		listArgs = append(listArgs, pageSize)
		lidx++
		if page > 0 {
			listSB.WriteString(fmt.Sprintf(" OFFSET $%d", lidx))
			listArgs = append(listArgs, (page-1)*pageSize)
			lidx++
		}
	}

	rows, err := conn.Query(ctx, listSB.String(), listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询数据库失败: %w", err)
	}
	defer rows.Close()

	idMap := make(map[int64]*Alert)
	ids := make([]int64, 0)
	for rows.Next() {
		a := &Alert{Label: make(map[string]string), Annotation: make(map[string]string)}
		if err := rows.Scan(&a.ID, &a.Status, &a.StartsAt, &a.EndsAt); err != nil {
			return nil, 0, fmt.Errorf("读取数据失败: %w", err)
		}
		alerts = append(alerts, a)
		idMap[a.ID] = a
		ids = append(ids, a.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("读取数据失败: %w", err)
	}

	if len(ids) == 0 {
		return alerts, int(total), nil
	}

	lrows, err := conn.Query(ctx, "SELECT alertid, label, value FROM alertlabel WHERE alertid = ANY($1)", ids)
	if err != nil {
		return nil, 0, fmt.Errorf("查询数据库失败: %w", err)
	}
	for lrows.Next() {
		var id int64
		var k, v string
		if err := lrows.Scan(&id, &k, &v); err != nil {
			lrows.Close()
			return nil, 0, fmt.Errorf("读取数据失败: %w", err)
		}
		if al, ok := idMap[id]; ok {
			al.Label[k] = v
		}
	}
	if err := lrows.Err(); err != nil {
		lrows.Close()
		return nil, 0, fmt.Errorf("读取数据失败: %w", err)
	}
	lrows.Close()

	arows, err := conn.Query(ctx, "SELECT alertid, annotation, value FROM alertannotation WHERE alertid = ANY($1)", ids)
	if err != nil {
		return nil, 0, fmt.Errorf("查询数据库失败: %w", err)
	}
	for arows.Next() {
		var id int64
		var k, v string
		if err := arows.Scan(&id, &k, &v); err != nil {
			arows.Close()
			return nil, 0, fmt.Errorf("读取数据失败: %w", err)
		}
		if al, ok := idMap[id]; ok {
			al.Annotation[k] = v
		}
	}
	if err := arows.Err(); err != nil {
		arows.Close()
		return nil, 0, fmt.Errorf("读取数据失败: %w", err)
	}
	arows.Close()

	return alerts, int(total), nil
}
