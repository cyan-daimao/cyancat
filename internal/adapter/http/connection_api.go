// Package http 提供 Wails 暴露给前端的 ConnectionAPI
package http

import (
	"cyancat/internal/adapter/dto"
	"cyancat/internal/application/connectionservice"
	"cyancat/internal/application/sqlcompleteservice"
	"cyancat/internal/infra/api"
)

// ConnectionAPI 连接管理 API（通过 Wails Bindings 暴露给前端）
type ConnectionAPI struct {
	svc            connectionservice.ConnectionService
	sqlCompleteSvc sqlcompleteservice.Service
}

// NewConnectionAPI 创建 ConnectionAPI
func NewConnectionAPI(svc connectionservice.ConnectionService, sqlCompleteSvc sqlcompleteservice.Service) *ConnectionAPI {
	return &ConnectionAPI{svc: svc, sqlCompleteSvc: sqlCompleteSvc}
}

// List 列出连接
func (a *ConnectionAPI) List(query *dto.ListConnectionRequest) *api.Response[[]*dto.ConnectionDTO] {
	if query == nil {
		query = &dto.ListConnectionRequest{}
	}

	list, err := a.svc.List(dto.ToListConnectionQuery(query))
	if err != nil {
		return api.Fail[[]*dto.ConnectionDTO](api.ErrorCode, nil, err.Error())
	}

	return api.Success(dto.ToConnectionDTOs(list))
}

// Page 分页查询连接
func (a *ConnectionAPI) Page(query *dto.PageConnectionRequest) *api.Response[*api.Page[*dto.ConnectionDTO]] {
	if query == nil {
		query = &dto.PageConnectionRequest{}
	}

	page, err := a.svc.Page(dto.ToPageConnectionQuery(query))
	if err != nil {
		return api.Fail[*api.Page[*dto.ConnectionDTO]](api.ErrorCode, nil, err.Error())
	}

	return api.Success(dto.ToConnectionPageDTO(page))
}

// GetByID 按 ID 查询连接
func (a *ConnectionAPI) GetByID(id int64) *api.Response[*dto.ConnectionDTO] {
	bo, err := a.svc.GetByID(id)
	if err != nil {
		return api.Fail[*dto.ConnectionDTO](api.ErrorCode, nil, err.Error())
	}
	return api.Success(dto.ToConnectionDTO(bo))
}

// Create 创建连接
func (a *ConnectionAPI) Create(req *dto.CreateConnectionRequest) *api.Response[*dto.ConnectionDTO] {
	if req == nil {
		return api.Fail[*dto.ConnectionDTO](api.BadRequestCode, nil, "request cannot be nil")
	}

	bo, err := a.svc.Create(dto.ToCreateConnectionCmd(req))
	if err != nil {
		return api.Fail[*dto.ConnectionDTO](api.ErrorCode, nil, err.Error())
	}

	return api.Success(dto.ToConnectionDTO(bo))
}

// Update 更新连接
func (a *ConnectionAPI) Update(id int64, req *dto.UpdateConnectionRequest) *api.Response[*dto.ConnectionDTO] {
	if req == nil {
		return api.Fail[*dto.ConnectionDTO](api.BadRequestCode, nil, "request cannot be nil")
	}

	bo, err := a.svc.Update(id, dto.ToUpdateConnectionCmd(req))
	if err != nil {
		return api.Fail[*dto.ConnectionDTO](api.ErrorCode, nil, err.Error())
	}

	return api.Success(dto.ToConnectionDTO(bo))
}

// Delete 删除连接
func (a *ConnectionAPI) Delete(id int64) *api.Response[bool] {
	if id <= 0 {
		return api.Fail[bool](api.BadRequestCode, false, "id must be positive")
	}

	if err := a.svc.Delete(&connectionservice.DeleteConnectionCmd{ID: id}); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}

	return api.Success(true)
}

// Test 测试连接（返回 TestConnectionResultDTO）
func (a *ConnectionAPI) Test(req *dto.TestConnectionRequest) *api.Response[*dto.TestConnectionResultDTO] {
	if req == nil {
		return api.Fail[*dto.TestConnectionResultDTO](api.BadRequestCode, nil, "request cannot be nil")
	}

	result, err := a.svc.Test(dto.ToTestConnectionCmd(req))
	if err != nil {
		// 即使 err 不为空，result 仍可能携带 Success=false + Message，给前端友好提示
		if result != nil {
			return api.Success(dto.ToTestConnectionResultDTO(result))
		}
		return api.Fail[*dto.TestConnectionResultDTO](api.ErrorCode, nil, err.Error())
	}

	return api.Success(dto.ToTestConnectionResultDTO(result))
}

// Open 打开已保存的连接（建立长连接，存入 SessionManager）
func (a *ConnectionAPI) Open(id int64) *api.Response[*dto.ConnectionDTO] {
	if id <= 0 {
		return api.Fail[*dto.ConnectionDTO](api.BadRequestCode, nil, "id must be positive")
	}
	bo, err := a.svc.Open(id)
	if err != nil {
		return api.Fail[*dto.ConnectionDTO](api.ErrorCode, nil, err.Error())
	}
	if a.sqlCompleteSvc != nil {
		a.sqlCompleteSvc.PrefetchConnectionCache(bo.ID, string(bo.Type), bo.Database)
	}
	return api.Success(dto.ToConnectionDTO(bo))
}

// Close 关闭已打开的连接
func (a *ConnectionAPI) Close(id int64) *api.Response[bool] {
	if id <= 0 {
		return api.Fail[bool](api.BadRequestCode, false, "id must be positive")
	}
	if err := a.svc.Close(id); err != nil {
		return api.Fail[bool](api.ErrorCode, false, err.Error())
	}
	return api.Success(true)
}
