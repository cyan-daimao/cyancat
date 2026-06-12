// Package mysql 提供 MySQL 驱动实现
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cyancat/internal/infra/driver"

	// 引入 MySQL 驱动注册到 database/sql
	_ "github.com/go-sql-driver/mysql"
)

// Driver MySQL 驱动实现
type Driver struct {
	dialect driver.Dialect
}

// New 创建 MySQL 驱动实例
func New() *Driver {
	return &Driver{
		dialect: &mysqlDialect{},
	}
}

// Type 返回驱动类型
func (d *Driver) Type() driver.DriverType {
	return driver.MySQL
}

// Dialect 返回 MySQL 方言
func (d *Driver) Dialect() driver.Dialect {
	return d.dialect
}

// Connect 建立 MySQL 连接
func (d *Driver) Connect(ctx context.Context, cfg driver.ConnConfig) (driver.Conn, error) {
	if cfg.Type != driver.MySQL {
		return nil, fmt.Errorf("mysql: expect type %q, got %q", driver.MySQL, cfg.Type)
	}

	dsn := buildDSN(cfg)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	connectCtx := ctx
	if cfg.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
		defer cancel()
	}

	if err := db.PingContext(connectCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: ping: %w", err)
	}

	conn := &Conn{db: db}
	conn.inspector = newInspector(conn)
	conn.ddl = newDDLGenerator(conn)
	return conn, nil
}

func buildDSN(cfg driver.ConnConfig) string {
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 3306
	}

	// user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=true
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.User, cfg.Password, host, port, cfg.Database)

	if cfg.SSL {
		dsn += "&tls=true"
	}

	for k, v := range cfg.Params {
		dsn += "&" + k + "=" + v
	}

	return dsn
}

// Conn MySQL 连接实现
type Conn struct {
	db        *sql.DB
	inspector driver.Inspector
	ddl       driver.DDLGenerator
}

// Ping 测试连接
func (c *Conn) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// WithDatabase MySQL 在同一连接内通过 USE 切换 database，返回当前连接即可。
func (c *Conn) WithDatabase(ctx context.Context, database string) (driver.Conn, func(), error) {
	return c, func() {}, nil
}

// Close 关闭连接
func (c *Conn) Close() error {
	return c.db.Close()
}

// Inspector 返回元数据查询器
func (c *Conn) Inspector() driver.Inspector {
	return c.inspector
}

// DDL 返回 DDL 生成器
func (c *Conn) DDL() driver.DDLGenerator {
	return c.ddl
}

// ServerVersion 返回 MySQL 服务端版本
func (c *Conn) ServerVersion(ctx context.Context) (string, error) {
	var v string
	if err := c.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&v); err != nil {
		return "", err
	}
	return v, nil
}

// Execute 执行 SQL（自动判断是查询还是 DML/DDL）
func (c *Conn) Execute(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	if isQuery(sqlText) {
		return c.executeQuery(ctx, sqlText, args...)
	}
	return c.executeNonQuery(ctx, sqlText, args...)
}

func (c *Conn) executeQuery(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	rows, err := c.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql: query: %w", err)
	}
	defer rows.Close()

	cols, err := buildColumns(rows)
	if err != nil {
		return nil, err
	}

	data, err := scanAll(rows, len(cols))
	if err != nil {
		return nil, err
	}

	return &driver.Result{
		Columns: cols,
		Rows:    data,
	}, nil
}

func (c *Conn) executeNonQuery(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	res, err := c.db.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql: exec: %w", err)
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return &driver.Result{
		RowsAffected: affected,
		LastInsertID: lastID,
	}, nil
}

// Stream 流式执行（返回游标）
func (c *Conn) Stream(ctx context.Context, sqlText string, args ...any) (driver.RowStream, error) {
	rows, err := c.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mysql: stream: %w", err)
	}

	cols, err := buildColumns(rows)
	if err != nil {
		_ = rows.Close()
		return nil, err
	}

	return &rowStream{rows: rows, cols: cols}, nil
}

// rowStream 把 *sql.Rows 包装成 driver.RowStream
type rowStream struct {
	rows *sql.Rows
	cols []driver.Column
	err  error
}

// Columns 列定义
func (s *rowStream) Columns() []driver.Column {
	return s.cols
}

// Next 移动到下一行
func (s *rowStream) Next() bool {
	return s.rows.Next()
}

// Scan 扫描当前行
func (s *rowStream) Scan() ([]any, error) {
	values, err := scanRow(s.rows, len(s.cols))
	if err != nil {
		s.err = err
		return nil, err
	}
	return values, nil
}

// Err 错误
func (s *rowStream) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.rows.Err()
}

// Close 关闭
func (s *rowStream) Close() error {
	return s.rows.Close()
}

// --- 通用辅助 ---

// isQuery 简单判断 SQL 是查询还是 DML/DDL（按首关键字）
func isQuery(sqlText string) bool {
	for i := 0; i < len(sqlText); i++ {
		c := sqlText[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		// 取前 6 个非空字符做前缀比较
		end := i + 6
		if end > len(sqlText) {
			end = len(sqlText)
		}
		prefix := sqlText[i:end]
		switch {
		case startsWithFold(prefix, "SELECT"),
			startsWithFold(prefix, "SHOW"),
			startsWithFold(prefix, "DESC"),
			startsWithFold(prefix, "EXPLAI"),
			startsWithFold(prefix, "WITH"):
			return true
		}
		return false
	}
	return false
}

func startsWithFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		a := s[i]
		b := prefix[i]
		if a >= 'a' && a <= 'z' {
			a -= 32
		}
		if b >= 'a' && b <= 'z' {
			b -= 32
		}
		if a != b {
			return false
		}
	}
	return true
}

// buildColumns 从 sql.Rows 元数据构建 driver.Column 列表
func buildColumns(rows *sql.Rows) ([]driver.Column, error) {
	names, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("mysql: columns: %w", err)
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("mysql: column types: %w", err)
	}

	cols := make([]driver.Column, len(names))
	for i, name := range names {
		col := driver.Column{Name: name}
		if i < len(types) {
			col.DatabaseType = types[i].DatabaseTypeName()
			if nullable, ok := types[i].Nullable(); ok {
				col.Nullable = nullable
			} else {
				col.Nullable = true
			}
		} else {
			col.Nullable = true
		}
		cols[i] = col
	}
	return cols, nil
}

// scanAll 把所有行 scan 到 [][]any
func scanAll(rows *sql.Rows, colCount int) ([][]any, error) {
	var result [][]any
	for rows.Next() {
		row, err := scanRow(rows, colCount)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// scanRow scan 单行；统一把 []byte 转 string，便于 JSON 序列化
func scanRow(rows *sql.Rows, colCount int) ([]any, error) {
	values := make([]any, colCount)
	scanPtrs := make([]any, colCount)
	for i := range values {
		scanPtrs[i] = &values[i]
	}
	if err := rows.Scan(scanPtrs...); err != nil {
		return nil, fmt.Errorf("mysql: scan: %w", err)
	}
	// 把 []byte 转 string，方便前端展示
	for i, v := range values {
		if b, ok := v.([]byte); ok {
			values[i] = string(b)
		}
	}
	return values, nil
}
