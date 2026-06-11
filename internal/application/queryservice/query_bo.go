package queryservice

import "time"

// ColumnBO 结果列定义
type ColumnBO struct {
	// Name 列名
	Name string
	// DatabaseType 数据库原生类型
	DatabaseType string
	// Nullable 是否可空
	Nullable bool
}

// QueryResultBO 查询结果业务对象
type QueryResultBO struct {
	// ConnID 连接 ID
	ConnID int64
	// SQL 实际执行的 SQL（多语句时为最后一条）
	SQL string
	// Columns 列定义
	Columns []ColumnBO
	// Rows 结果数据（每行 []any）
	Rows [][]any
	// RowsAffected 受影响行数
	RowsAffected int64
	// LastInsertID 最后插入 ID
	LastInsertID int64
	// Duration 执行耗时
	Duration time.Duration
	// Truncated 结果是否被默认 LIMIT 截断
	Truncated bool
}

// QueryHistoryBO 查询历史业务对象
type QueryHistoryBO struct {
	// ID 主键
	ID int64
	// ConnID 连接 ID
	ConnID int64
	// Database 数据库名
	Database string
	// SQL 执行的 SQL
	SQL string
	// Status 执行状态（success/error）
	Status string
	// ErrorMessage 错误信息
	ErrorMessage string
	// RowCount 返回行数
	RowCount int64
	// DurationMs 耗时（毫秒）
	DurationMs int64
	// ExecutedAt 执行时间
	ExecutedAt time.Time
}

// HistoryRepository 查询历史仓储接口
type HistoryRepository interface {
	// Save 保存一条历史
	Save(bo *QueryHistoryBO) error
	// Page 分页查询历史
	Page(query *HistoryQuery) ([]*QueryHistoryBO, int64, error)
}
