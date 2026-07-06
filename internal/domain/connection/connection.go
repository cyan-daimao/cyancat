// Package connection 定义数据库连接的领域模型与业务规则。
//
// Connection 是核心领域对象，表示一个数据库连接配置。
// 采用充血模型：业务校验、状态变化、持久化行为均由领域对象自身承担。
package connection

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cyancat/internal/infra/driver"
)

// 连接分组常量
const (
	// GroupProduction 生产环境
	GroupProduction = "production"
	// GroupStaging 预发布环境
	GroupStaging = "staging"
	// GroupDevelopment 开发环境
	GroupDevelopment = "development"
	// GroupTest 测试环境
	GroupTest = "test"
)

// 合法的分组集合
var validGroups = map[string]bool{
	GroupProduction:  true,
	GroupStaging:     true,
	GroupDevelopment: true,
	GroupTest:        true,
}

// Connection 数据库连接领域实体
type Connection struct {
	// ID 主键
	ID int64
	// Name 连接名称
	Name string
	// Type 驱动类型（mysql / postgres / sqlite）
	Type driver.DriverType
	// Host 主机地址
	Host string
	// Port 端口号
	Port int
	// User 用户名
	User string
	// Password 密码（加密存储）
	Password string
	// Database 默认数据库
	Database string
	// SSL 是否启用 SSL
	SSL bool
	// Group 连接分组
	Group string
	// Color 标记颜色（用于 UI 显示）
	Color string
	// LastConnectedAt 最后连接时间
	LastConnectedAt *time.Time
	// CreatedAt 创建时间
	CreatedAt time.Time
	// UpdatedAt 更新时间
	UpdatedAt time.Time
	// DeletedAt 删除时间（软删除）
	DeletedAt *time.Time
}

// Save 保存连接（新建），执行归一化和校验后委托 Repository 持久化
func (c *Connection) Save(repo Repository) error {
	if repo == nil {
		return errors.New("connection: repository cannot be nil")
	}
	if err := c.prepareSave(); err != nil {
		return err
	}
	return repo.Save(c)
}

// Update 更新连接，执行归一化和校验后委托 Repository 持久化
func (c *Connection) Update(repo Repository) error {
	if repo == nil {
		return errors.New("connection: repository cannot be nil")
	}
	if c == nil {
		return errors.New("connection: cannot be nil")
	}
	if c.ID <= 0 {
		return errors.New("connection: id must be positive")
	}
	if err := c.prepareUpdate(); err != nil {
		return err
	}
	return repo.Update(c)
}

// Delete 删除连接（软删除）
func (c *Connection) Delete(repo Repository) error {
	if repo == nil {
		return errors.New("connection: repository cannot be nil")
	}
	if c == nil {
		return errors.New("connection: cannot be nil")
	}
	if c.ID <= 0 {
		return errors.New("connection: id must be positive")
	}
	now := time.Now()
	c.DeletedAt = &now
	return repo.Delete(c.ID)
}

// TouchConnected 更新最后连接时间
func (c *Connection) TouchConnected() error {
	if c == nil {
		return errors.New("connection: cannot be nil")
	}
	now := time.Now()
	c.LastConnectedAt = &now
	return nil
}

// SetPassword 设置密码（后续由 keychain 加密存储，Domain 层只持有明文临时值）
func (c *Connection) SetPassword(password string) {
	if c == nil {
		return
	}
	c.Password = password
}

// --- 私有方法 ---

func (c *Connection) prepareSave() error {
	c.normalize()
	return c.validate()
}

func (c *Connection) prepareUpdate() error {
	c.normalize()
	return c.validate()
}

func (c *Connection) normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.Host = strings.TrimSpace(c.Host)
	c.User = strings.TrimSpace(c.User)
	c.Database = strings.TrimSpace(c.Database)
	c.Group = strings.TrimSpace(c.Group)
	c.Color = strings.TrimSpace(c.Color)

	// 默认分组
	if c.Group == "" {
		c.Group = GroupDevelopment
	}

	// 默认端口
	if c.Port <= 0 {
		switch c.Type {
		case driver.MySQL:
			c.Port = 3306
		case driver.PostgreSQL:
			c.Port = 5432
		case driver.SQLite:
			c.Port = 0
		case driver.StarRocks:
			c.Port = 9030
		case driver.Kafka:
			c.Port = 9092
		case driver.Redis:
			c.Port = 6379
		}
	}

	// Kafka / Redis 默认主机
	if (c.Type == driver.Kafka || c.Type == driver.Redis) && c.Host == "" {
		c.Host = "127.0.0.1"
	}

	// 默认主机
	if c.Host == "" && c.Type != driver.SQLite && c.Type != driver.Kafka && c.Type != driver.Redis {
		c.Host = "127.0.0.1"
	}
}

func (c *Connection) validate() error {
	if c.Name == "" {
		return errors.New("connection: name is required")
	}
	if !c.Type.IsValid() {
		return fmt.Errorf("connection: invalid driver type %q", c.Type)
	}
	if c.Host == "" {
		if c.Type == driver.SQLite {
			return errors.New("connection: sqlite database file is required")
		}
		return errors.New("connection: host is required")
	}
	if c.Type != driver.SQLite {
		if c.Port <= 0 || c.Port > 65535 {
			return fmt.Errorf("connection: port must be between 1 and 65535, got %d", c.Port)
		}
		if c.Type != driver.Kafka && c.Type != driver.Redis && c.User == "" {
			return errors.New("connection: user is required")
		}
	}
	if c.Group != "" && !validGroups[c.Group] {
		return fmt.Errorf("connection: invalid group %q", c.Group)
	}
	return nil
}
