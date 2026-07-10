package parser

import (
	"github.com/pganalyze/pg_query_go/v5"
)

// Postgres PostgreSQL 方言解析器。
type Postgres struct{}

// NewPostgres 创建 PostgreSQL 解析器。
func NewPostgres() *Postgres {
	return &Postgres{}
}

// Parse 解析 SQL 并返回补全上下文。
func (p *Postgres) Parse(sql string, cursorLine, cursorColumn int) (*ParseResult, error) {
	res := &ParseResult{
		Context: CtxUnknown,
		Tables:  []TableRef{},
		CTEs:    []string{},
	}

	offset := cursorOffset(sql, cursorLine, cursorColumn)
	if offset < 0 || offset > len(sql) {
		offset = len(sql)
	}
	prefix := sql[:offset]

	// 1. 判断 member access
	memberAccess := detectMemberAccess(prefix)
	if memberAccess != "" {
		res.Context = CtxColumn
		res.TablePrefix = memberAccess
		res.IsMemberAccess = true
	}

	// 2. 使用 pg_query_go 解析
	parsed, err := pg_query.Parse(sql)
	if err == nil && parsed != nil && len(parsed.Stmts) > 0 {
		// 只提取光标所在语句的表，避免多语句相同 alias 串表
		if raw := findPostgresStmtAtCursor(parsed.Stmts, offset); raw != nil && raw.Stmt != nil {
			extractPostgresNode(raw.Stmt, res)
		}
	}

	// 3. fallback
	if len(res.Tables) == 0 {
		start, _ := statementBounds(sql, offset)
		fallbackExtractTables(prefix[start:], res)
	}

	// 4. 推断上下文
	if !res.IsMemberAccess {
		res.Context = inferContext(prefix)
	}

	return res, nil
}

// extractPostgresNode 递归遍历 PostgreSQL AST 节点。
func extractPostgresNode(node *pg_query.Node, res *ParseResult) {
	if node == nil {
		return
	}

	switch n := node.Node.(type) {
	case *pg_query.Node_SelectStmt:
		extractPostgresSelectStmt(n.SelectStmt, res)
	case *pg_query.Node_RangeVar:
		extractPostgresRangeVar(n.RangeVar, res)
	case *pg_query.Node_JoinExpr:
		extractPostgresJoin(n.JoinExpr, res)
	case *pg_query.Node_RangeSubselect:
		if n.RangeSubselect != nil && n.RangeSubselect.Alias != nil {
			ref := TableRef{Name: "", Alias: n.RangeSubselect.Alias.Aliasname, IsCTE: false}
			res.Tables = append(res.Tables, ref)
		}
	}
}

func extractPostgresJoin(join *pg_query.JoinExpr, res *ParseResult) {
	if join == nil {
		return
	}
	extractPostgresNode(join.Larg, res)
	extractPostgresNode(join.Rarg, res)
}

// findPostgresStmtAtCursor 根据 StmtLocation/StmtLen 找到包含光标偏移的语句。
func findPostgresStmtAtCursor(stmts []*pg_query.RawStmt, offset int) *pg_query.RawStmt {
	for _, raw := range stmts {
		if raw == nil {
			continue
		}
		start := int(raw.StmtLocation)
		end := start + int(raw.StmtLen)
		if offset >= start && offset <= end {
			return raw
		}
	}
	// 兜底：返回最后一条语句
	for i := len(stmts) - 1; i >= 0; i-- {
		if stmts[i] != nil {
			return stmts[i]
		}
	}
	return nil
}

func extractPostgresSelectStmt(stmt *pg_query.SelectStmt, res *ParseResult) {
	if stmt == nil {
		return
	}
	for _, f := range stmt.FromClause {
		extractPostgresNode(f, res)
	}
	// PostgreSQL 的 join 在 FromClause 中作为 RangeSubselect/JoinExpr 出现
	// CTEs
	if stmt.WithClause != nil {
		for _, cte := range stmt.WithClause.Ctes {
			extractPostgresCTE(cte, res)
		}
	}
}

func extractPostgresRangeVar(rv *pg_query.RangeVar, res *ParseResult) {
	if rv == nil {
		return
	}
	ref := TableRef{Name: rv.Relname}
	if rv.Schemaname != "" {
		ref.Schema = rv.Schemaname
	}
	if rv.Alias != nil {
		ref.Alias = rv.Alias.Aliasname
	}
	res.Tables = append(res.Tables, ref)
}

func extractPostgresCTE(node *pg_query.Node, res *ParseResult) {
	if node == nil {
		return
	}
	cte, ok := node.Node.(*pg_query.Node_CommonTableExpr)
	if !ok || cte.CommonTableExpr == nil {
		return
	}
	if cte.CommonTableExpr.Ctename != "" {
		res.CTEs = append(res.CTEs, cte.CommonTableExpr.Ctename)
	}
}
