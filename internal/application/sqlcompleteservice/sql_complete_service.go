// Package sqlcompleteservice 提供 SQL 编辑器的智能补全服务。
package sqlcompleteservice

// Service SQL 补全服务接口。
type Service interface {
	// Complete 根据当前 SQL 和光标位置返回补全候选。
	Complete(query *CompleteQuery) (*CompleteResult, error)
}
