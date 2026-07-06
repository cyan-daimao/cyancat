// Package dto 定义 adapter 层的请求/响应对象和转换函数
package dto

import (
	"time"

	"cyancat/internal/application/connectionservice"
	"cyancat/internal/infra/api"
	"cyancat/internal/infra/driver"
)

// --- 请求对象 ---

// CreateConnectionRequest 创建连接请求
type CreateConnectionRequest struct {
	// Name 连接名称
	Name string `json:"name" binding:"required"`
	// Type 驱动类型（mysql / postgres / sqlite / starrocks）
	Type string `json:"type" binding:"required"`
	// Host 主机地址
	Host string `json:"host" binding:"required"`
	// Port 端口号
	Port int `json:"port"`
	// User 用户名
	User string `json:"user" binding:"required"`
	// Password 密码
	Password string `json:"password"`
	// Database 默认数据库
	Database string `json:"database"`
	// SSL 是否启用 SSL
	SSL bool `json:"ssl"`
	// Group 连接分组
	Group string `json:"group"`
	// Color 标记颜色
	Color string `json:"color"`
}

// UpdateConnectionRequest 更新连接请求
type UpdateConnectionRequest struct {
	// Name 连接名称
	Name string `json:"name" binding:"required"`
	// Type 驱动类型（mysql / postgres / sqlite / starrocks）
	Type string `json:"type" binding:"required"`
	// Host 主机地址
	Host string `json:"host" binding:"required"`
	// Port 端口号
	Port int `json:"port"`
	// User 用户名
	User string `json:"user" binding:"required"`
	// Password 密码（空字符串表示不修改）
	Password string `json:"password"`
	// Database 默认数据库
	Database string `json:"database"`
	// SSL 是否启用 SSL
	SSL bool `json:"ssl"`
	// Group 连接分组
	Group string `json:"group"`
	// Color 标记颜色
	Color string `json:"color"`
}

// TestConnectionRequest 测试连接请求
type TestConnectionRequest struct {
	// Type 驱动类型（mysql / postgres / sqlite / starrocks）
	Type string `json:"type" binding:"required"`
	// Host 主机地址
	Host string `json:"host" binding:"required"`
	// Port 端口号
	Port int `json:"port"`
	// User 用户名
	User string `json:"user" binding:"required"`
	// Password 密码
	Password string `json:"password"`
	// Database 默认数据库
	Database string `json:"database"`
	// SSL 是否启用 SSL
	SSL bool `json:"ssl"`
}

// ListConnectionRequest 列表查询请求
type ListConnectionRequest struct {
	// Group 按分组过滤
	Group string `json:"group"`
	// Type 按驱动类型过滤
	Type string `json:"type"`
	// Keyword 按名称/主机关键字模糊匹配
	Keyword string `json:"keyword"`
}

// PageConnectionRequest 分页查询请求
type PageConnectionRequest struct {
	// Group 按分组过滤
	Group string `json:"group"`
	// Type 按驱动类型过滤
	Type string `json:"type"`
	// Keyword 按名称/主机关键字模糊匹配
	Keyword string `json:"keyword"`
	// Page 页码（从 1 开始）
	Page int `json:"page"`
	// PageSize 每页条数
	PageSize int `json:"pageSize"`
}

// --- 响应对象 ---

// ConnectionDTO 连接响应对象（不暴露密码）
type ConnectionDTO struct {
	// ID 主键
	ID int64 `json:"id"`
	// Name 连接名称
	Name string `json:"name"`
	// Type 驱动类型
	Type string `json:"type"`
	// Host 主机地址
	Host string `json:"host"`
	// Port 端口号
	Port int `json:"port"`
	// User 用户名
	User string `json:"user"`
	// Database 默认数据库
	Database string `json:"database"`
	// SSL 是否启用 SSL
	SSL bool `json:"ssl"`
	// Group 连接分组
	Group string `json:"group"`
	// Color 标记颜色
	Color string `json:"color"`
	// LastConnectedAt 最后连接时间
	LastConnectedAt *time.Time `json:"lastConnectedAt"`
	// CreatedAt 创建时间
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updatedAt"`
}

// --- 转换函数 ---

// ToConnectionDTO 把 BO 转换为 DTO（脱敏：不暴露密码）
func ToConnectionDTO(bo *connectionservice.ConnectionBO) *ConnectionDTO {
	if bo == nil {
		return nil
	}
	return &ConnectionDTO{
		ID:              bo.ID,
		Name:            bo.Name,
		Type:            bo.Type.String(),
		Host:            bo.Host,
		Port:            bo.Port,
		User:            bo.User,
		Database:        bo.Database,
		SSL:             bo.SSL,
		Group:           bo.Group,
		Color:           bo.Color,
		LastConnectedAt: bo.LastConnectedAt,
		CreatedAt:       bo.CreatedAt,
		UpdatedAt:       bo.UpdatedAt,
	}
}

// ToConnectionPageDTO 把 BO 分页结构转为 DTO 分页结构
func ToConnectionPageDTO(p *api.Page[*connectionservice.ConnectionBO]) *api.Page[*ConnectionDTO] {
	if p == nil {
		return nil
	}
	return &api.Page[*ConnectionDTO]{
		Page:     p.Page,
		PageSize: p.PageSize,
		Total:    p.Total,
		List:     ToConnectionDTOs(p.List),
	}
}

// ToConnectionDTOs 批量转换
func ToConnectionDTOs(bos []*connectionservice.ConnectionBO) []*ConnectionDTO {
	if len(bos) == 0 {
		return make([]*ConnectionDTO, 0)
	}
	result := make([]*ConnectionDTO, 0, len(bos))
	for _, bo := range bos {
		result = append(result, ToConnectionDTO(bo))
	}
	return result
}

// ToCreateConnectionCmd 把请求转为创建命令
func ToCreateConnectionCmd(req *CreateConnectionRequest) *connectionservice.CreateConnectionCmd {
	if req == nil {
		return nil
	}
	return &connectionservice.CreateConnectionCmd{
		Name:     req.Name,
		Type:     driver.DriverType(req.Type),
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSL:      req.SSL,
		Group:    req.Group,
		Color:    req.Color,
	}
}

// ToUpdateConnectionCmd 把请求转为更新命令
func ToUpdateConnectionCmd(req *UpdateConnectionRequest) *connectionservice.UpdateConnectionCmd {
	if req == nil {
		return nil
	}
	return &connectionservice.UpdateConnectionCmd{
		Name:     req.Name,
		Type:     driver.DriverType(req.Type),
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSL:      req.SSL,
		Group:    req.Group,
		Color:    req.Color,
	}
}

// ToTestConnectionCmd 把请求转为测试命令
func ToTestConnectionCmd(req *TestConnectionRequest) *connectionservice.TestConnectionCmd {
	if req == nil {
		return nil
	}
	return &connectionservice.TestConnectionCmd{
		Type:     driver.DriverType(req.Type),
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
		SSL:      req.SSL,
	}
}

// ToListConnectionQuery 把请求转为列表查询
func ToListConnectionQuery(req *ListConnectionRequest) *connectionservice.ListConnectionQuery {
	if req == nil {
		return &connectionservice.ListConnectionQuery{}
	}
	return &connectionservice.ListConnectionQuery{
		Group:   req.Group,
		Type:    req.Type,
		Keyword: req.Keyword,
	}
}

// ToPageConnectionQuery 把请求转为分页查询
func ToPageConnectionQuery(req *PageConnectionRequest) *connectionservice.PageConnectionQuery {
	if req == nil {
		return &connectionservice.PageConnectionQuery{}
	}
	return &connectionservice.PageConnectionQuery{
		Group:    req.Group,
		Type:     req.Type,
		Keyword:  req.Keyword,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
}

// ToDeleteConnectionCmd 把 ID 转为删除命令
func ToDeleteConnectionCmd(id int64) *connectionservice.DeleteConnectionCmd {
	return &connectionservice.DeleteConnectionCmd{ID: id}
}

// TestConnectionResultDTO 测试连接结果响应
type TestConnectionResultDTO struct {
	// Success 是否成功
	Success bool `json:"success"`
	// Message 提示信息
	Message string `json:"message"`
	// ServerVersion 数据库服务端版本
	ServerVersion string `json:"serverVersion"`
}

// ToTestConnectionResultDTO BO -> DTO
func ToTestConnectionResultDTO(bo *connectionservice.TestConnectionResult) *TestConnectionResultDTO {
	if bo == nil {
		return nil
	}
	return &TestConnectionResultDTO{
		Success:       bo.Success,
		Message:       bo.Message,
		ServerVersion: bo.ServerVersion,
	}
}
