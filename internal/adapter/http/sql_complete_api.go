package http

import (
	"cyancat/internal/adapter/dto"
	"cyancat/internal/application/sqlcompleteservice"
	"cyancat/internal/infra/api"
)

// SqlCompleteAPI SQL 补全 API（暴露给前端）
type SqlCompleteAPI struct {
	svc sqlcompleteservice.Service
}

// NewSqlCompleteAPI 创建 SqlCompleteAPI
func NewSqlCompleteAPI(svc sqlcompleteservice.Service) *SqlCompleteAPI {
	return &SqlCompleteAPI{svc: svc}
}

// Complete SQL 智能补全
func (a *SqlCompleteAPI) Complete(req *dto.SqlCompleteRequest) *api.Response[[]dto.SqlCompleteCandidate] {
	if req == nil {
		return api.Fail[[]dto.SqlCompleteCandidate](api.BadRequestCode, nil, "request cannot be nil")
	}
	result, err := a.svc.Complete(dto.ToCompleteQuery(req))
	if err != nil {
		return api.Fail[[]dto.SqlCompleteCandidate](api.ErrorCode, nil, err.Error())
	}
	candidates := dto.ToSqlCompleteCandidates(result.Candidates)
	return api.Success(candidates)
}
