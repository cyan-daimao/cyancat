// Package sqlcompleteservice 提供 SQL 编辑器的智能补全服务。
package sqlcompleteservice

// Service SQL 补全服务接口。
type Service interface {
	// Complete 根据当前 SQL 和光标位置返回补全候选。
	Complete(query *CompleteQuery) (*CompleteResult, error)
	// ClearTableCache 清除指定范围的表名缓存。
	// database/schema 为空时清除该 connID 下的所有表名缓存。
	ClearTableCache(connID int64, database, schema string)
	// ClearSchemaCache 清除指定范围的 schema 缓存。database 为空时清除该 connID 下的所有 schema 缓存。
	ClearSchemaCache(connID int64, database string)
	// PrefetchConnectionCache 连接打开后后台预缓存数据库/模式/表名。
	PrefetchConnectionCache(connID int64, connectionType, database string)
}
