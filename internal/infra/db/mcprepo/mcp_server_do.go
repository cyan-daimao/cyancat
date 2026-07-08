// Package mcprepo 提供 MCP Server 配置的 GORM 持久化实现
package mcprepo

import "time"

// McpServerDO MCP Server 配置持久层对象
type McpServerDO struct {
	// ID 主键
	ID int64 `gorm:"column:id; primaryKey; autoIncrement; comment:'主键id'"`
	// ConnID 连接 ID
	ConnID int64 `gorm:"column:conn_id; not null; uniqueIndex:uk_conn_id; comment:'连接ID'"`
	// Enabled 是否已启用
	Enabled bool `gorm:"column:enabled; not null; default:false; comment:'是否已启用'"`
	// AllowSelect 允许 SELECT
	AllowSelect bool `gorm:"column:allow_select; not null; default:false; comment:'允许SELECT'"`
	// AllowInsert 允许 INSERT
	AllowInsert bool `gorm:"column:allow_insert; not null; default:false; comment:'允许INSERT'"`
	// AllowUpdate 允许 UPDATE
	AllowUpdate bool `gorm:"column:allow_update; not null; default:false; comment:'允许UPDATE'"`
	// AllowDelete 允许 DELETE
	AllowDelete bool `gorm:"column:allow_delete; not null; default:false; comment:'允许DELETE'"`
	// AllowDDL 允许 DDL
	AllowDDL bool `gorm:"column:allow_ddl; not null; default:false; comment:'允许DDL'"`
	// Port 监听端口
	Port int `gorm:"column:port; comment:'监听端口'"`
	// Token 访问令牌
	Token string `gorm:"column:token; size:256; comment:'访问令牌'"`
	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"column:created_at; not null; default:CURRENT_TIMESTAMP; comment:'创建时间'"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `gorm:"column:updated_at; not null; default:CURRENT_TIMESTAMP; comment:'更新时间'"`
	// DeletedAt 删除时间（软删除）
	DeletedAt *time.Time `gorm:"column:deleted_at; comment:'删除时间'"`
}

// TableName 表名
func (McpServerDO) TableName() string {
	return "mcp_server"
}
