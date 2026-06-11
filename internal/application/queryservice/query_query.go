package queryservice

import "time"

// HistoryQuery 查询历史的查询条件
type HistoryQuery struct {
	// ConnID 按连接过滤（0 表示不过滤）
	ConnID int64
	// Keyword 按 SQL 关键字过滤
	Keyword string
	// Status 按执行状态过滤（success/error）
	Status string
	// StartTime 起始时间
	StartTime *time.Time
	// EndTime 结束时间
	EndTime *time.Time
	// Page 页码（从 1 开始）
	Page int
	// PageSize 每页条数
	PageSize int
}
