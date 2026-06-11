package queryservice

// ExecuteCmd SQL 执行命令
type ExecuteCmd struct {
	// ConnID 目标连接 ID
	ConnID int64
	// SQL 待执行 SQL（可能含多语句）
	SQL string
	// Stream 是否走流式（true 时即使小结果集也通过 EventBus 推送）
	Stream bool
	// MaxRows 同步返回的最大行数（0 表示使用默认值 1000）
	MaxRows int
	// Database 执行前切换到的数据库
	Database string
	// Schema 执行前切换到的 schema
	Schema string
}
