package dto

import (
	"time"

	"cyancat/internal/application/queryservice"
)

// --- 请求 ---

// ExecuteQueryRequest 执行查询请求
type ExecuteQueryRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID" binding:"required"`
	// SQL 待执行 SQL
	SQL string `json:"sql" binding:"required"`
	// Stream 是否流式
	Stream bool `json:"stream"`
	// MaxRows 最大行数（默认 1000）
	MaxRows int `json:"maxRows"`
	// Database 执行前切换到的数据库
	Database string `json:"database"`
	// Schema 执行前切换到的 schema
	Schema string `json:"schema"`
}

// QueryHistoryRequest 查询历史请求
type QueryHistoryRequest struct {
	// ConnID 按连接过滤
	ConnID int64 `json:"connID"`
	// Keyword 关键字
	Keyword string `json:"keyword"`
	// Status 状态
	Status string `json:"status"`
	// Page 页码
	Page int `json:"page"`
	// PageSize 每页条数
	PageSize int `json:"pageSize"`
}

// --- 响应 ---

// ColumnDTO 结果列
type ColumnDTO struct {
	// Name 列名
	Name string `json:"name"`
	// DatabaseType 数据库类型
	DatabaseType string `json:"databaseType"`
	// Nullable 是否可空
	Nullable bool `json:"nullable"`
}

// QueryResultDTO 查询结果
type QueryResultDTO struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// SQL 实际执行的 SQL
	SQL string `json:"sql"`
	// Columns 列定义
	Columns []ColumnDTO `json:"columns"`
	// Rows 结果数据
	Rows [][]any `json:"rows"`
	// RowsAffected 受影响行数
	RowsAffected int64 `json:"rowsAffected"`
	// LastInsertID 最后插入 ID
	LastInsertID int64 `json:"lastInsertID"`
	// DurationMs 耗时（毫秒）
	DurationMs int64 `json:"durationMs"`
	// Truncated 是否被截断
	Truncated bool `json:"truncated"`
}

// QueryHistoryDTO 查询历史
type QueryHistoryDTO struct {
	// ID 主键
	ID int64 `json:"id"`
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// SQL 执行 SQL
	SQL string `json:"sql"`
	// Status 状态
	Status string `json:"status"`
	// ErrorMessage 错误信息
	ErrorMessage string `json:"errorMessage"`
	// RowCount 返回行数
	RowCount int64 `json:"rowCount"`
	// DurationMs 耗时
	DurationMs int64 `json:"durationMs"`
	// ExecutedAt 执行时间
	ExecutedAt time.Time `json:"executedAt"`
}

// --- 转换 ---

// ToColumnDTOs BO -> DTO
func ToColumnDTOs(cols []queryservice.ColumnBO) []ColumnDTO {
	if len(cols) == 0 {
		return make([]ColumnDTO, 0)
	}
	result := make([]ColumnDTO, 0, len(cols))
	for _, c := range cols {
		result = append(result, ColumnDTO{
			Name:         c.Name,
			DatabaseType: c.DatabaseType,
			Nullable:     c.Nullable,
		})
	}
	return result
}

// ToQueryResultDTO BO -> DTO
func ToQueryResultDTO(bo *queryservice.QueryResultBO) *QueryResultDTO {
	if bo == nil {
		return nil
	}
	return &QueryResultDTO{
		ConnID:       bo.ConnID,
		SQL:          bo.SQL,
		Columns:      ToColumnDTOs(bo.Columns),
		Rows:         bo.Rows,
		RowsAffected: bo.RowsAffected,
		LastInsertID: bo.LastInsertID,
		DurationMs:   bo.Duration.Milliseconds(),
		Truncated:    bo.Truncated,
	}
}

// ToExecuteQueryCmd Request -> Cmd
func ToExecuteQueryCmd(req *ExecuteQueryRequest) *queryservice.ExecuteCmd {
	if req == nil {
		return nil
	}
	return &queryservice.ExecuteCmd{
		ConnID:   req.ConnID,
		SQL:      req.SQL,
		Stream:   req.Stream,
		MaxRows:  req.MaxRows,
		Database: req.Database,
		Schema:   req.Schema,
	}
}

// ToHistoryQuery Request -> Query
func ToHistoryQuery(req *QueryHistoryRequest) *queryservice.HistoryQuery {
	if req == nil {
		return &queryservice.HistoryQuery{}
	}
	return &queryservice.HistoryQuery{
		ConnID:   req.ConnID,
		Keyword:  req.Keyword,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
}

// ToQueryHistoryDTO BO -> DTO
func ToQueryHistoryDTO(bo *queryservice.QueryHistoryBO) *QueryHistoryDTO {
	if bo == nil {
		return nil
	}
	return &QueryHistoryDTO{
		ID:           bo.ID,
		ConnID:       bo.ConnID,
		SQL:          bo.SQL,
		Status:       bo.Status,
		ErrorMessage: bo.ErrorMessage,
		RowCount:     bo.RowCount,
		DurationMs:   bo.DurationMs,
		ExecutedAt:   bo.ExecutedAt,
	}
}

// ToQueryHistoryDTOs 批量
func ToQueryHistoryDTOs(list []*queryservice.QueryHistoryBO) []*QueryHistoryDTO {
	if len(list) == 0 {
		return make([]*QueryHistoryDTO, 0)
	}
	result := make([]*QueryHistoryDTO, 0, len(list))
	for _, bo := range list {
		result = append(result, ToQueryHistoryDTO(bo))
	}
	return result
}
