package connectionservice

import (
	"time"

	"cyancat/internal/infra/driver"
)

// ConnectionBO 连接业务对象（application 层返回和组装对象）
type ConnectionBO struct {
	// ID 主键
	ID int64
	// Name 连接名称
	Name string
	// Type 驱动类型
	Type driver.DriverType
	// Host 主机地址
	Host string
	// Port 端口号
	Port int
	// User 用户名
	User string
	// Password 密码（仅在内部使用，不暴露给前端）
	Password string
	// Database 默认数据库
	Database string
	// SSL 是否启用 SSL
	SSL bool
	// Group 连接分组
	Group string
	// Color 标记颜色
	Color string
	// LastConnectedAt 最后连接时间
	LastConnectedAt *time.Time
	// CreatedAt 创建时间
	CreatedAt time.Time
	// UpdatedAt 更新时间
	UpdatedAt time.Time
}

// TestConnectionResult 测试连接结果
type TestConnectionResult struct {
	// Success 是否成功
	Success bool
	// Message 提示信息
	Message string
	// ServerVersion 数据库版本
	ServerVersion string
}
