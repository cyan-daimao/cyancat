// Package historyrepo 提供查询历史的 GORM 持久化实现
package historyrepo

import (
	"time"
)

// QueryHistoryDO 查询历史持久层对象
type QueryHistoryDO struct {
	// ID 主键
	ID int64 `gorm:"column:id; primaryKey; autoIncrement; comment:'主键id'"`
	// ConnID 连接 ID
	ConnID int64 `gorm:"column:conn_id; not null; index:idx_conn_id; comment:'连接 ID'"`
	// Database 数据库名
	Database string `gorm:"column:database; size:128; comment:'数据库名'"`
	// SQL 执行的 SQL
	SQL string `gorm:"column:sql; type:text; not null; comment:'执行的 SQL'"`
	// Status 执行状态（success / error）
	Status string `gorm:"column:status; size:16; not null; default:'success'; comment:'执行状态'"`
	// ErrorMessage 错误信息
	ErrorMessage string `gorm:"column:error_message; size:1024; comment:'错误信息'"`
	// RowCount 返回行数
	RowCount int64 `gorm:"column:row_count; comment:'返回行数'"`
	// Duration 执行耗时（毫秒）
	Duration int64 `gorm:"column:duration; comment:'执行耗时(ms)'"`
	// CreatedAt 创建时间
	CreatedAt time.Time `gorm:"column:created_at; not null; default:CURRENT_TIMESTAMP; comment:'创建时间'"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `gorm:"column:updated_at; not null; default:CURRENT_TIMESTAMP; comment:'更新时间'"`
	// DeletedAt 删除时间（软删除）
	DeletedAt *time.Time `gorm:"column:deleted_at; comment:'删除时间'"`
}

// TableName 表名
func (QueryHistoryDO) TableName() string {
	return "query_history"
}
