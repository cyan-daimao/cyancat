package parser

import (
	"strings"
)

// New 根据连接类型创建对应方言的解析器。
func New(connectionType string) Parser {
	switch strings.ToLower(connectionType) {
	case "mysql", "mariadb", "starrocks":
		return NewMySQL()
	case "postgres", "postgresql":
		return NewPostgres()
	default:
		// SQLite 等先用 MySQL parser 做 best-effort 解析
		return NewMySQL()
	}
}
