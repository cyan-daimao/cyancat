// Package session 管理"已激活的数据库连接"的运行时容器
//
// 注意：Session 是基础设施概念（运行时连接池），不是领域概念。
// 它的职责是：按 connID 索引活跃的 driver.Conn 实例，提供 Open/Get/Close。
package session

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cyancat/internal/infra/driver"
)

// Manager 活动连接管理器
type Manager interface {
	// Open 用 ConnConfig 建立连接并存入容器
	Open(ctx context.Context, connID int64, cfg driver.ConnConfig) error
	// Get 获取已激活的 Conn（如未激活返回错误）
	Get(connID int64) (driver.Conn, error)
	// DriverType 返回已激活连接的驱动类型
	DriverType(connID int64) (driver.DriverType, error)
	// IsOpen 检查连接是否已激活
	IsOpen(connID int64) bool
	// Close 关闭指定连接
	Close(connID int64) error
	// CloseAll 关闭所有连接
	CloseAll() error
	// List 列出所有活跃连接的 ID
	List() []int64
}

// entry 单条会话条目
type entry struct {
	// conn 活跃的 driver.Conn
	conn driver.Conn
	// driverType 驱动类型
	driverType driver.DriverType
	// openedAt 打开时间
	openedAt time.Time
}

// managerImpl 内存实现：map + sync.RWMutex
type managerImpl struct {
	mu       sync.RWMutex
	sessions map[int64]*entry
}

// NewManager 创建会话管理器
func NewManager() Manager {
	return &managerImpl{
		sessions: make(map[int64]*entry),
	}
}

// Open 建立连接并存入容器（如已有则在新连接成功后再关闭旧的）
func (m *managerImpl) Open(ctx context.Context, connID int64, cfg driver.ConnConfig) error {
	if connID <= 0 {
		return errors.New("session: connID must be positive")
	}

	// 通过 Registry 拿 Driver
	d, err := driver.Get(cfg.Type)
	if err != nil {
		return fmt.Errorf("session: get driver: %w", err)
	}

	// 建立新连接
	conn, err := d.Connect(ctx, cfg)
	if err != nil {
		return fmt.Errorf("session: connect: %w", err)
	}

	// 新连接已建立，现在再关闭旧的
	m.mu.Lock()
	existing, ok := m.sessions[connID]
	if ok {
		_ = existing.conn.Close()
	}
	m.sessions[connID] = &entry{
		conn:       conn,
		driverType: cfg.Type,
		openedAt:   time.Now(),
	}
	m.mu.Unlock()
	return nil
}

// Get 获取已激活的 Conn
func (m *managerImpl) Get(connID int64) (driver.Conn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.sessions[connID]
	if !ok {
		return nil, fmt.Errorf("session: connection %d not open", connID)
	}
	return e.conn, nil
}

// DriverType 返回已激活连接的驱动类型
func (m *managerImpl) DriverType(connID int64) (driver.DriverType, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.sessions[connID]
	if !ok {
		return "", fmt.Errorf("session: connection %d not open", connID)
	}
	return e.driverType, nil
}

// IsOpen 检查连接是否已激活
func (m *managerImpl) IsOpen(connID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.sessions[connID]
	return ok
}

// Close 关闭指定连接
func (m *managerImpl) Close(connID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.sessions[connID]
	if !ok {
		return nil
	}
	delete(m.sessions, connID)
	return e.conn.Close()
}

// CloseAll 关闭所有连接
func (m *managerImpl) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var firstErr error
	for id, e := range m.sessions {
		if err := e.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(m.sessions, id)
	}
	return firstErr
}

// List 列出所有活跃连接 ID
func (m *managerImpl) List() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]int64, 0, len(m.sessions))
	for id := range m.sessions {
		ids = append(ids, id)
	}
	return ids
}
