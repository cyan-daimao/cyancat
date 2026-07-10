package parser

import (
	"regexp"
)

// statementBounds 返回光标所在语句的字节边界 [start, end)。
// 扫描时跳过字符串字面量（'...'、"..."、`...`）和注释（--、/*...*/），
// 避免把字符串/注释里的分号误判为语句分隔符。
func statementBounds(sql string, offset int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(sql) {
		offset = len(sql)
	}

	start := 0
	for i := offset - 1; i >= 0; i-- {
		if sql[i] == ';' && !isInsideLiteralOrComment(sql, i) {
			start = i + 1
			break
		}
	}
	// 跳过语句开头的空白
	for start < len(sql) {
		c := sql[start]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			start++
			continue
		}
		break
	}

	end := len(sql)
	for i := offset; i < len(sql); i++ {
		if sql[i] == ';' && !isInsideLiteralOrComment(sql, i) {
			end = i
			break
		}
	}

	return start, end
}

// isInsideLiteralOrComment 判断 offset 位置是否位于字符串字面量或注释内部。
// 简单正向扫描实现，足够应对常规 SQL 分号定位。
func isInsideLiteralOrComment(sql string, offset int) bool {
	var inSingle, inDouble, inBacktick bool
	var inLineComment, inBlockComment bool
	for i := 0; i < offset; i++ {
		if inLineComment {
			if sql[i] == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			if i+1 < len(sql) && sql[i] == '*' && sql[i+1] == '/' {
				inBlockComment = false
				i++
			}
			continue
		}
		if inSingle {
			if sql[i] == '\'' {
				// 处理转义 ''
				if i+1 < len(sql) && sql[i+1] == '\'' {
					i++
					continue
				}
				inSingle = false
			}
			continue
		}
		if inDouble {
			if sql[i] == '"' {
				if i+1 < len(sql) && sql[i+1] == '"' {
					i++
					continue
				}
				inDouble = false
			}
			continue
		}
		if inBacktick {
			if sql[i] == '`' {
				if i+1 < len(sql) && sql[i+1] == '`' {
					i++
					continue
				}
				inBacktick = false
			}
			continue
		}
		if i+1 < len(sql) {
			if sql[i] == '-' && sql[i+1] == '-' {
				inLineComment = true
				i++
				continue
			}
			if sql[i] == '/' && sql[i+1] == '*' {
				inBlockComment = true
				i++
				continue
			}
		}
		if sql[i] == '\'' {
			inSingle = true
			continue
		}
		if sql[i] == '"' {
			inDouble = true
			continue
		}
		if sql[i] == '`' {
			inBacktick = true
			continue
		}
	}
	return inSingle || inDouble || inBacktick || inLineComment || inBlockComment
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
