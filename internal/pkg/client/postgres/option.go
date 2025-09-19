package postgres

import (
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

// Option 使用函数式选项模式配置连接池。
type Option func(cfg *pgxpool.Config)

// WithPoolConfig 直接对底层配置进行修改，提供最大灵活性。
func WithPoolConfig(fn func(cfg *pgxpool.Config)) Option {
    return func(cfg *pgxpool.Config) { fn(cfg) }
}

// WithMaxConns 设置最大连接数。
func WithMaxConns(n int32) Option {
    return func(cfg *pgxpool.Config) { cfg.MaxConns = n }
}

// WithMinConns 设置最小保持的空闲连接数。
func WithMinConns(n int32) Option {
    return func(cfg *pgxpool.Config) { cfg.MinConns = n }
}

// WithMaxConnLifetime 设置连接的最长生命周期。
func WithMaxConnLifetime(d time.Duration) Option {
    return func(cfg *pgxpool.Config) { cfg.MaxConnLifetime = d }
}

// WithMaxConnIdleTime 设置连接的最长空闲时间。
func WithMaxConnIdleTime(d time.Duration) Option {
    return func(cfg *pgxpool.Config) { cfg.MaxConnIdleTime = d }
}

// WithHealthCheckPeriod 设置健康检查间隔。
func WithHealthCheckPeriod(d time.Duration) Option {
    return func(cfg *pgxpool.Config) { cfg.HealthCheckPeriod = d }
}

