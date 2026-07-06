// Package starrocks 提供 StarRocks 驱动实现。
//
// StarRocks 使用 MySQL 协议，因此直接复用 internal/infra/driver/mysql 的连接与
// 方言；Inspector 与 DDL 单独实现以支持 catalog -> database -> table 三层结构。
package starrocks

import (
	"context"
	"database/sql"
	"fmt"

	"cyancat/internal/infra/driver"
	mysqldriver "cyancat/internal/infra/driver/mysql"
)

// defaultPort StarRocks 默认查询端口（MySQL 协议）
const defaultPort = 9030

// Driver StarRocks 驱动实现
type Driver struct {
	mysql driver.Driver
}

// New 创建 StarRocks 驱动实例
func New() *Driver {
	return &Driver{
		mysql: mysqldriver.New(),
	}
}

// Type 返回驱动类型
func (d *Driver) Type() driver.DriverType {
	return driver.StarRocks
}

// Dialect 返回 MySQL 方言（StarRocks 标识符引用、占位符与 MySQL 一致）
func (d *Driver) Dialect() driver.Dialect {
	return d.mysql.Dialect()
}

// Connect 建立 StarRocks 连接
func (d *Driver) Connect(ctx context.Context, cfg driver.ConnConfig) (driver.Conn, error) {
	if cfg.Type != driver.StarRocks {
		return nil, fmt.Errorf("starrocks: expect type %q, got %q", driver.StarRocks, cfg.Type)
	}

	if cfg.Port == 0 {
		cfg.Port = defaultPort
	}

	// StarRocks 兼容 MySQL 协议，使用 MySQL 驱动建立实际连接
	mysqlCfg := cfg
	mysqlCfg.Type = driver.MySQL
	mysqlConn, err := d.mysql.Connect(ctx, mysqlCfg)
	if err != nil {
		return nil, err
	}

	c := &conn{
		Conn:      mysqlConn,
		inspector: newInspector(mysqlConn),
		ddl:       newDDLGenerator(mysqlConn),
	}
	return c, nil
}

// conn 包装 mysql.Conn，替换 Inspector 与 DDL 为 StarRocks 专用实现。
type conn struct {
	driver.Conn
	inspector driver.Inspector
	ddl       driver.DDLGenerator
}

// Inspector 返回 StarRocks 元数据查询器
func (c *conn) Inspector() driver.Inspector {
	return c.inspector
}

// DDL 返回 StarRocks DDL 生成器
func (c *conn) DDL() driver.DDLGenerator {
	return c.ddl
}

// db 从底层 mysql.Conn 拿到 *sql.DB
func db(c driver.Conn) *sql.DB {
	mysqlConn, ok := c.(*mysqldriver.Conn)
	if !ok {
		return nil
	}
	return mysqlConn.DB()
}
