package connectionservice

import "cyancat/internal/infra/driver"

// CreateConnectionCmd 创建连接命令
type CreateConnectionCmd struct {
	// Name 连接名称
	Name string
	// Type 驱动类型（mysql / postgres / sqlite）
	Type driver.DriverType
	// Host 主机地址
	Host string
	// Port 端口号
	Port int
	// User 用户名
	User string
	// Password 密码
	Password string
	// Database 默认数据库
	Database string
	// SSL 是否启用 SSL
	SSL bool
	// Group 连接分组
	Group string
	// Color 标记颜色
	Color string
}

// UpdateConnectionCmd 更新连接命令
type UpdateConnectionCmd struct {
	// Name 连接名称
	Name string
	// Type 驱动类型（mysql / postgres / sqlite）
	Type driver.DriverType
	// Host 主机地址
	Host string
	// Port 端口号
	Port int
	// User 用户名
	User string
	// Password 密码（空字符串表示不修改）
	Password string
	// Database 默认数据库
	Database string
	// SSL 是否启用 SSL
	SSL bool
	// Group 连接分组
	Group string
	// Color 标记颜色
	Color string
}

// DeleteConnectionCmd 删除连接命令
type DeleteConnectionCmd struct {
	// ID 连接 ID
	ID int64
}

// TestConnectionCmd 测试连接命令
type TestConnectionCmd struct {
	// Type 驱动类型
	Type driver.DriverType
	// Host 主机地址
	Host string
	// Port 端口号
	Port int
	// User 用户名
	User string
	// Password 密码
	Password string
	// Database 默认数据库
	Database string
	// SSL 是否启用 SSL
	SSL bool
}
