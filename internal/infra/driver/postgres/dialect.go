package postgres

// postgresDialect PostgreSQL 方言：双引号引号 + $N 占位符
type postgresDialect struct{}

// QuoteIdent 双引号包裹标识符
func (d *postgresDialect) QuoteIdent(ident string) string {
	out := make([]byte, 0, len(ident)+2)
	out = append(out, '"')
	for i := 0; i < len(ident); i++ {
		if ident[i] == '"' {
			out = append(out, '"', '"')
		} else {
			out = append(out, ident[i])
		}
	}
	out = append(out, '"')
	return string(out)
}

// Placeholder PostgreSQL 用 $1, $2, ... 占位
func (d *postgresDialect) Placeholder(n int) string {
	// $1 ~ $99
	if n < 10 {
		return string([]byte{'$', byte('0' + n)})
	}
	return string([]byte{'$', byte('0' + n/10), byte('0' + n%10)})
}

// DefaultLimit 默认 LIMIT 1000
func (d *postgresDialect) DefaultLimit() int {
	return 1000
}