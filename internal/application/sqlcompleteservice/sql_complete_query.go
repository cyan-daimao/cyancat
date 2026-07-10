package sqlcompleteservice

// CompleteQuery SQL 补全请求。
type CompleteQuery struct {
	ConnID         int64
	ConnectionType string // mysql | postgres | sqlite | starrocks
	Database       string
	Schema         string
	SQL            string
	CursorLine     int // 1-based
	CursorColumn   int // 1-based
	Prefix         string // 当前光标处已输入的标识符前缀（用于表名搜索）
}
