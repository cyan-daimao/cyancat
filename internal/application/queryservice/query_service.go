// Package queryservice 提供 SQL 查询执行的应用层服务
package queryservice

import (
	"cyancat/internal/infra/api"
)

// QueryService SQL 查询执行服务接口
type QueryService interface {
	// Execute 执行 SQL（小结果集同步返回；大结果集自动流式推送 EventBus）
	Execute(cmd *ExecuteCmd) (*QueryResultBO, error)
	// Cancel 取消正在执行的查询（V1.5 扩展）
	Cancel(connID int64) error
	// History 查询历史
	History(query *HistoryQuery) (*api.Page[*QueryHistoryBO], error)
}
