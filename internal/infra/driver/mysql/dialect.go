package mysql

import "strconv"

// mysqlDialect MySQL 方言：反引号引号 + ? 占位符
type mysqlDialect struct{}

// QuoteIdent 反引号包裹标识符
func (d *mysqlDialect) QuoteIdent(ident string) string {
	// 简单转义内部反引号
	out := make([]byte, 0, len(ident)+2)
	out = append(out, '`')
	for i := 0; i < len(ident); i++ {
		if ident[i] == '`' {
			out = append(out, '`', '`')
		} else {
			out = append(out, ident[i])
		}
	}
	out = append(out, '`')
	return string(out)
}

// Placeholder MySQL 始终用 ?
func (d *mysqlDialect) Placeholder(n int) string {
	_ = n
	return "?"
}

// DefaultLimit 默认 LIMIT 1000
func (d *mysqlDialect) DefaultLimit() int {
	return 1000
}

// _ 保留 strconv 引用以备扩展（如未来需要拼 LIMIT 偏移）
var _ = strconv.Itoa
