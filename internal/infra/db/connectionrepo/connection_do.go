// Package connectionrepo 提供 Connection 仓储的 GORM 实现
package connectionrepo

import (
	"time"
)

// ConnectionDO 数据库连接持久层对象
type ConnectionDO struct {
	// ID 主键
	ID int64 `gorm:"column:id; primaryKey; autoIncrement; comment:'主键id'"`
	// Name 连接名称
	Name string `gorm:"column:name; not null; size:128; uniqueIndex:uk_name_deleted; comment:'连接名称'"`
	// Type 驱动类型（mysql / postgres）
	Type string `gorm:"column:type; not null; size:32; index:idx_type; comment:'驱动类型'"`
	// Host 主机地址
	Host string `gorm:"column:host; not null; size:255; comment:'主机地址'"`
	// Port 端口号
	Port int `gorm:"column:port; not null; comment:'端口号'"`
	// User 用户名
	User string `gorm:"column:user; not null; size:128; comment:'用户名'"`
	// PasswordEncrypted 加密后的密码密文
	PasswordEncrypted string `gorm:"column:password_encrypted; size:2048; comment:'加密后的密码密文'"`
	// Database 默认数据库
	Database string `gorm:"column:database; size:128; comment:'默认数据库'"`
	// SSL 是否启用 SSL
	SSL bool `gorm:"column:ssl; not null; default:false; comment:'是否启用 SSL'"`
	// Group 连接分组
	Group string `gorm:"column:group; not null; size:32; default:'development'; index:idx_group; comment:'连接分组'"`
	// Color 标记颜色
	Color string `gorm:"column:color; size:16; comment:'标记颜色'"`
	// LastConnectedAt 最后连接时间
	LastConnectedAt *time.Time `gorm:"column:last_connected_at; comment:'最后连接时间'"`
	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"column:created_at; not null; default:CURRENT_TIMESTAMP; comment:'创建时间'"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `gorm:"column:updated_at; not null; default:CURRENT_TIMESTAMP; comment:'更新时间'"`
	// DeletedAt 删除时间（软删除）
	DeletedAt *time.Time `gorm:"column:deleted_at; uniqueIndex:uk_name_deleted; comment:'删除时间'"`
	// CreatedBy 创建人
	CreatedBy string `gorm:"column:created_by; size:64; comment:'创建人'"`
	// UpdatedBy 更新人
	UpdatedBy string `gorm:"column:updated_by; size:64; comment:'更新人'"`
}

// TableName 表名
func (ConnectionDO) TableName() string {
	return "connection"
}
