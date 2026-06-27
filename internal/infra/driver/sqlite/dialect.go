package sqlite

// sqliteDialect SQLite 方言：双引号标识符 + ? 占位符
type sqliteDialect struct{}

func (d *sqliteDialect) QuoteIdent(ident string) string {
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

func (d *sqliteDialect) Placeholder(n int) string {
	_ = n
	return "?"
}

func (d *sqliteDialect) DefaultLimit() int {
	return 1000
}
