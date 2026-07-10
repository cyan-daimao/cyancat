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
		for _, stmt := range stmts {
			extractTablesFromStmt(stmt, res)
		}
	}

	// 3. 如果 parser 失败或没有提取到表，用 fallback 正则提取
	if len(res.Tables) == 0 {
		fallbackExtractTables(prefix, res)
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

// cursorOffset 将 1-based 行列转换为字节偏移。
func cursorOffset(sql string, line, column int) int {
	if line <= 1 {
		if column <= 1 {
			return 0
		}
		return min(column-1, len(sql))
	}
	currentLine := 1
	for i, c := range []byte(sql) {
		if c == '\n' {
			currentLine++
			if currentLine == line {
				next := i + column
				if next > len(sql) {
					return len(sql)
				}
				return next
			}
		}
	}
	return len(sql)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractTablesFromStmt 从语句中提取表引用。
func extractTablesFromStmt(stmt ast.StmtNode, res *ParseResult) {
	v := &tableVisitor{res: res}
	stmt.Accept(v)
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

// fallbackExtractTables 用正则从 FROM/JOIN 子句提取表名。
func fallbackExtractTables(prefix string, res *ParseResult) {
	re := regexp.MustCompile(`(?i)(?:from|join)\s+(?:` + "`" + `?([A-Za-z_][A-Za-z0-9_]*)` + "`" + `?)(?:\s+(?:as\s+)?(?:` + "`" + `?([A-Za-z_][A-Za-z0-9_]*)` + "`" + `?))?`)
	matches := re.FindAllStringSubmatch(prefix, -1)
	for _, m := range matches {
		ref := TableRef{Name: unquote(m[1])}
		if len(m) > 2 && m[2] != "" {
			ref.Alias = unquote(m[2])
		}
		res.Tables = append(res.Tables, ref)
	}
}

func unquote(s string) string {
	if len(s) >= 2 && ((s[0] == '`' && s[len(s)-1] == '`') || (s[0] == '"' && s[len(s)-1] == '"')) {
		return s[1 : len(s)-1]
	}
	return s
}
