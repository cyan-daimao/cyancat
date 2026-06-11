package http

import (
	"cyancat/internal/adapter/dto"
	"cyancat/internal/application/queryservice"
	"cyancat/internal/infra/api"
)

// QueryAPI SQL 查询 API（暴露给前端）
type QueryAPI struct {
	svc queryservice.QueryService
}

// NewQueryAPI 创建 QueryAPI
func NewQueryAPI(svc queryservice.QueryService) *QueryAPI {
	return &QueryAPI{svc: svc}
}

// Execute 执行 SQL
func (a *QueryAPI) Execute(req *dto.ExecuteQueryRequest) *api.Response[*dto.QueryResultDTO] {
	if req == nil {
		return api.Fail[*dto.QueryResultDTO](api.BadRequestCode, nil, "request cannot be nil")
	}

	bo, err := a.svc.Execute(dto.ToExecuteQueryCmd(req))
	if err != nil {
		return api.Fail[*dto.QueryResultDTO](api.ErrorCode, nil, err.Error())
	}

	return api.Success(dto.ToQueryResultDTO(bo))
}

// Cancel 取消查询
func (a *QueryAPI) Cancel(connID int64) *api.Response[bool] {
	if connID <= 0 {
		return api.Fail[bool](api.BadRequestCode, false, "connID must be positive")
	}
	if err := a.svc.Cancel(connID); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}
	return api.Success(true)
}

// History 查询历史
func (a *QueryAPI) History(req *dto.QueryHistoryRequest) *api.Response[*api.Page[*dto.QueryHistoryDTO]] {
	page, err := a.svc.History(dto.ToHistoryQuery(req))
	if err != nil {
		return api.Fail[*api.Page[*dto.QueryHistoryDTO]](api.ErrorCode, nil, err.Error())
	}

	dtoPage := &api.Page[*dto.QueryHistoryDTO]{
		Page:     page.Page,
		PageSize: page.PageSize,
		Total:    page.Total,
		List:     dto.ToQueryHistoryDTOs(page.List),
	}
	return api.Success(dtoPage)
}
