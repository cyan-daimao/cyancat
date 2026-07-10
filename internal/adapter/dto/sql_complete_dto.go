package dto

import "cyancat/internal/application/sqlcompleteservice"

// SqlCompleteRequest SQL 补全请求
type SqlCompleteRequest struct {
	// ConnID 连接 ID
	ConnID int64 `json:"connID"`
	// ConnectionType 连接类型：mysql | postgres | sqlite | starrocks
	ConnectionType string `json:"connectionType"`
	// Database 数据库名
	Database string `json:"database"`
	// Schema schema 名
	Schema string `json:"schema"`
	// SQL 当前编辑器完整内容
	SQL string `json:"sql"`
	// CursorLine 光标行号（1-based）
	CursorLine int `json:"cursorLine"`
	// CursorColumn 光标列号（1-based）
	CursorColumn int `json:"cursorColumn"`
}

// SqlCompleteCandidate 单个补全候选
type SqlCompleteCandidate struct {
	// Label 显示文本
	Label string `json:"label"`
	// Kind 类型：keyword | table | view | column | function
	Kind string `json:"kind"`
	// Detail 补充信息（字段类型/表注释等）
	Detail string `json:"detail"`
	// InsertText 实际插入文本
	InsertText string `json:"insertText"`
	// SortText 排序权重
	SortText string `json:"sortText"`
}

// ToCompleteQuery Request → Query
func ToCompleteQuery(req *SqlCompleteRequest) *sqlcompleteservice.CompleteQuery {
	if req == nil {
		return nil
	}
	return &sqlcompleteservice.CompleteQuery{
		ConnID:         req.ConnID,
		ConnectionType: req.ConnectionType,
		Database:       req.Database,
		Schema:         req.Schema,
		SQL:            req.SQL,
		CursorLine:     req.CursorLine,
		CursorColumn:   req.CursorColumn,
	}
}

// ToSqlCompleteCandidate BO → DTO
func ToSqlCompleteCandidate(bo sqlcompleteservice.CompleteCandidate) SqlCompleteCandidate {
	return SqlCompleteCandidate{
		Label:      bo.Label,
		Kind:       string(bo.Kind),
		Detail:     bo.Detail,
		InsertText: bo.InsertText,
		SortText:   bo.SortText,
	}
}

// ToSqlCompleteCandidates BOs → DTOs
func ToSqlCompleteCandidates(bos []sqlcompleteservice.CompleteCandidate) []SqlCompleteCandidate {
	if bos == nil {
		return make([]SqlCompleteCandidate, 0)
	}
	res := make([]SqlCompleteCandidate, 0, len(bos))
	for _, bo := range bos {
		res = append(res, ToSqlCompleteCandidate(bo))
	}
	return res
}
