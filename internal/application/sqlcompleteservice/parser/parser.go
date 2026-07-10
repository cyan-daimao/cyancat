// Package parser 提供按方言解析 SQL 并提取补全上下文的能力。
package parser

// TableRef 表示 SQL 中引用的表及其别名。
type TableRef struct {
	Name    string // 原始表名
	Alias   string // 别名，可能为空
	Schema  string // schema，可能为空
	IsCTE   bool   // 是否是 CTE
}

// CompletionContext 当前光标处的补全上下文。
type CompletionContext string

const (
	CtxUnknown    CompletionContext = "unknown"
	CtxColumn     CompletionContext = "column"     // 需要提示字段，可能有表前缀
	CtxTable      CompletionContext = "table"      // 需要提示表名
	CtxSchema     CompletionContext = "schema"     // 需要提示 schema
	CtxDatabase   CompletionContext = "database"   // 需要提示 database
	CtxKeyword    CompletionContext = "keyword"    // 需要提示关键字
	CtxFunction   CompletionContext = "function"   // 需要提示函数
)

// ParseResult SQL 解析结果。
type ParseResult struct {
	Context       CompletionContext
	Tables        []TableRef   // FROM/JOIN 中的表引用
	CTEs          []string     // CTE 名称
	TablePrefix   string       // 当前输入的表/别名前缀，如 "a." 中的 "a"
	IsMemberAccess bool        // 是否是 table.column / alias.column 形式
}

// Parser SQL 解析器接口。
type Parser interface {
	// Parse 解析 SQL 并返回补全上下文。
	// cursorLine 和 cursorColumn 均为 1-based。
	Parse(sql string, cursorLine, cursorColumn int) (*ParseResult, error)
}
