package parser

import (
	"regexp"
	"strings"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

// MySQL MySQL 方言解析器（兼容 MariaDB / StarRocks）。
type MySQL struct{}

// NewMySQL 创建 MySQL 解析器。
func NewMySQL() *MySQL {
	return &MySQL{}
}

// Parse 解析 SQL 并返回补全上下文。
func (m *MySQL) Parse(sql string, cursorLine, cursorColumn int) (*ParseResult, error) {
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

	// 1. 判断是否是 member access：table. 或 alias.
	memberAccess := detectMemberAccess(prefix)
	if memberAccess != "" {
		res.Context = CtxColumn
		res.TablePrefix = memberAccess
		res.IsMemberAccess = true
	}

	// 2. 尝试用 TiDB parser 解析完整 SQL，提取表引用
	p := parser.New()
	stmts, _, err := p.Parse(sql, "", "")
	if err == nil && len(stmts) > 0 {
		// 只提取光标所在语句的表，避免多语句相同 alias 串表
		if stmt := findStmtAtCursor(stmts, sql, offset); stmt != nil {
			extractTablesFromStmt(stmt, res)
		}
	}

	// 3. 如果 parser 失败或没有提取到表，用 fallback 正则提取（限制在当前语句内）
	if len(res.Tables) == 0 {
		start, _ := statementBounds(sql, offset)
		fallbackExtractTables(prefix[start:], res)
	}

	// 4. 如果没有识别为 member access，根据上下文判断
	if !res.IsMemberAccess {
		res.Context = inferContext(prefix)
	}

	return res, nil
}

// detectMemberAccess 判断光标前是否是 identifier. 形式。
func detectMemberAccess(prefix string) string {
	trimmed := strings.TrimRight(prefix, " \t\n\r")
	re := regexp.MustCompile(`(?:([A-Za-z_][A-Za-z0-9_]*)|` + "`" + `([^` + "`" + `]*)` + "`" + `|\["([^"]*)"\])\.\s*$`)
	matches := re.FindStringSubmatch(trimmed)
	if matches == nil {
		return ""
	}
	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			return matches[i]
		}
	}
	return ""
}

// inferContext 根据前缀判断补全上下文。
func inferContext(prefix string) CompletionContext {
	upper := strings.ToUpper(strings.TrimSpace(prefix))
	// 形如 FROM <partial> / JOIN <partial> / ,<partial> 应提示表名/模式名
	tokens := lastSignificantTokens(prefix, 2)
	if len(tokens) == 2 {
		switch tokens[0] {
		case "FROM", "JOIN", ",":
			return CtxTable
		}
	}
	lastToken := lastSignificantToken(prefix)
	// 简单启发式：FROM/JOIN 后提示表名
	if strings.Contains(upper, "FROM ") || strings.Contains(upper, "JOIN ") {
		if lastToken == "FROM" || lastToken == "JOIN" || lastToken == "," {
			return CtxTable
		}
	}
	// SELECT / WHERE / GROUP BY / ORDER BY / HAVING / ON 后提示字段
	if strings.HasPrefix(upper, "SELECT") || strings.Contains(upper, " WHERE ") ||
		strings.Contains(upper, " GROUP BY ") || strings.Contains(upper, " ORDER BY ") ||
		strings.Contains(upper, " HAVING ") || strings.Contains(upper, " ON ") {
		return CtxColumn
	}
	return CtxKeyword
}

// lastSignificantTokens 返回前缀中最后 n 个非空 token（按空格/逗号分割）。
func lastSignificantTokens(s string, n int) []string {
	re := regexp.MustCompile(`(?:\s+|,)+`)
	parts := re.Split(strings.TrimSpace(s), -1)
	tokens := make([]string, 0, n)
	for i := len(parts) - 1; i >= 0 && len(tokens) < n; i-- {
		if parts[i] != "" {
			tokens = append(tokens, strings.ToUpper(parts[i]))
		}
	}
	// 反转回正序
	for i, j := 0, len(tokens)-1; i < j; i, j = i+1, j-1 {
		tokens[i], tokens[j] = tokens[j], tokens[i]
	}
	return tokens
}

// lastSignificantToken 返回前缀中最后一个非空 token。
func lastSignificantToken(s string) string {
	re := regexp.MustCompile(`(?:\s+|,)+`)
	parts := re.Split(strings.TrimSpace(s), -1)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return strings.ToUpper(parts[i])
		}
	}
	return ""
}

// extractTablesFromStmt 从语句中提取表引用。
func extractTablesFromStmt(stmt ast.StmtNode, res *ParseResult) {
	v := &tableVisitor{res: res}
	stmt.Accept(v)
}

// findStmtAtCursor 根据 OriginTextPosition 找到包含光标偏移的语句。
func findStmtAtCursor(stmts []ast.StmtNode, sql string, offset int) ast.StmtNode {
	if len(stmts) == 0 {
		return nil
	}
	// 按语句起始偏移排序
	type posStmt struct {
		pos  int
		stmt ast.StmtNode
	}
	ordered := make([]posStmt, 0, len(stmts))
	for _, stmt := range stmts {
		pos := stmt.OriginTextPosition()
		ordered = append(ordered, posStmt{pos: pos, stmt: stmt})
	}
	for i := 0; i < len(ordered)-1; i++ {
		for j := i + 1; j < len(ordered); j++ {
			if ordered[j].pos < ordered[i].pos {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	for i := range ordered {
		start := ordered[i].pos
		end := len(sql)
		if i+1 < len(ordered) {
			end = ordered[i+1].pos
		}
		if offset >= start && offset < end {
			return ordered[i].stmt
		}
	}
	return ordered[len(ordered)-1].stmt
}

type tableVisitor struct {
	res *ParseResult
}

func (v *tableVisitor) Enter(in ast.Node) (ast.Node, bool) {
	switch n := in.(type) {
	case *ast.TableSource:
		if tn, ok := n.Source.(*ast.TableName); ok {
			ref := TableRef{
				Name:  tn.Name.String(),
				Alias: n.AsName.String(),
			}
			if tn.Schema.String() != "" {
				ref.Schema = tn.Schema.String()
			}
			v.res.Tables = append(v.res.Tables, ref)
		}
	case *ast.CommonTableExpression:
		if n.Name.String() != "" {
			v.res.CTEs = append(v.res.CTEs, n.Name.String())
		}
	}
	return in, false
}

func (v *tableVisitor) Leave(in ast.Node) (ast.Node, bool) {
	return in, true
}
