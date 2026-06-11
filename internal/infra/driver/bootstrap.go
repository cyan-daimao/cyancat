package driver

// Bootstrap 默认注册中心初始化函数，由各驱动包在 main.go 中显式注册：
//
//	import (
//	    "cyancat/internal/infra/driver"
//	    "cyancat/internal/infra/driver/mysql"
//	    "cyancat/internal/infra/driver/postgres"
//	)
//
//	func init() {
//	    driver.Register(mysql.New())
//	    driver.Register(postgres.New())
//	}
//
// 此处仅放置占位文档，避免循环依赖。