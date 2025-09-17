package slurmdb

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"golang.org/x/sync/singleflight"
)

// PoolKey 唯一标识 SlurmDB
type poolKey struct {
	ClusterName string         // 对应 Slurm.Conf ClusterName
	Address     netip.AddrPort // 集群所对应的 Slurmdb 地址
}

// String 输出格式为 <ClusterName>:IP:Port
func (pk poolKey) String() string {
	return fmt.Sprintf("%s:%s", pk.ClusterName, pk.Address)
}

func NewPoolKey(cluster, ip string, port uint16) (string, error) {
	var err error
	pk := poolKey{}
	if cluster == "" {
		return "", fmt.Errorf("参数 cluster 不能为空")
	}
	pk.ClusterName = cluster
	pk.Address, err = netip.ParseAddrPort(fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return "", fmt.Errorf("无法执行地址解析函数ParseAddrPort(%s): %w", fmt.Sprintf("%s:%d", ip, port), err)
	}
	return fmt.Sprintf("%s:%s", pk.ClusterName, pk.Address), nil
}

type SlurmDBPool struct {
	mu   sync.RWMutex
	g    singleflight.Group
	pool map[string]*SlurmDB
}

type Conf struct {
	Cluster  string // Slurm.Conf 中的 Cluster 名称
	IP       string
	Port     uint16
	Database string // 数据库名称
	User     string // 用户名
	Passwd   string // 密码
}

func (conf Conf) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?sslmode=false", conf.User, conf.Passwd, conf.IP, conf.Port, conf.Database)
}

// FetchOrCreateFetch 根据 conf 生成的 key 获取 pool 中的 SlurmDB；若不存在则创建并返回。
// 该函数为并发安全：
// - 使用读写锁保护对内部 map 的访问；
// - 使用 singleflight 保证同一 key 的创建只会执行一次。
func (p *SlurmDBPool) FetchOrCreateFetch(ctx context.Context, conf Conf) (*SlurmDB, error) {
	key, err := NewPoolKey(conf.Cluster, conf.IP, conf.Port)
	if err != nil {
		return nil, fmt.Errorf("无法创建 PoolKey: %w", err)
	}

	// 快路径：已存在则直接返回
	p.mu.RLock()
	if db, ok := p.pool[key]; ok && db != nil {
		p.mu.RUnlock()
		return db, nil
	}
	p.mu.RUnlock()

	// 使用 singleflight 保证并发下同一 key 仅创建一次
	v, err, _ := p.g.Do(key, func() (any, error) {
		// 双检，避免等待期间已被其他协程创建
		p.mu.RLock()
		if db, ok := p.pool[key]; ok && db != nil {
			p.mu.RUnlock()
			return db, nil
		}
		p.mu.RUnlock()

		// 实例化 SlurmDB（通过 conf.DSN() 创建连接）
		newDB, err := NewSlurmDB(ctx, conf)
		if err != nil {
			return nil, err
		}

		// 存入池
		p.mu.Lock()
		if p.pool == nil {
			p.pool = make(map[string]*SlurmDB)
		}
		p.pool[key] = newDB
		p.mu.Unlock()

		return newDB, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*SlurmDB), nil
}
