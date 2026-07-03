// Package postgres 提供 PostgreSQL 驱动实现
package postgres

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"cyancat/internal/infra/driver"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Driver PostgreSQL 驱动实现
type Driver struct {
	dialect driver.Dialect
}

// New 创建 PostgreSQL 驱动实例
func New() *Driver {
	return &Driver{
		dialect: &postgresDialect{},
	}
}

// Type 返回驱动类型
func (d *Driver) Type() driver.DriverType {
	return driver.PostgreSQL
}

// Dialect 返回 PostgreSQL 方言
func (d *Driver) Dialect() driver.Dialect {
	return d.dialect
}

// Connect 建立 PostgreSQL 连接（使用 pgxpool）
func (d *Driver) Connect(ctx context.Context, cfg driver.ConnConfig) (driver.Conn, error) {
	if cfg.Type != driver.PostgreSQL {
		return nil, fmt.Errorf("postgres: expect type %q, got %q", driver.PostgreSQL, cfg.Type)
	}

	cfg.Database = normalizeDatabase(cfg.Database)
	dsn := buildDSN(cfg)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	poolCfg.MaxConns = 1
	poolCfg.MinConns = 1
	poolCfg.MaxConnLifetime = time.Hour

	connectCtx := ctx
	if cfg.ConnectTimeout > 0 {
		var cancel context.CancelFunc
		connectCtx, cancel = context.WithTimeout(ctx, cfg.ConnectTimeout)
		defer cancel()
	}

	pool, err := pgxpool.NewWithConfig(connectCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: new pool: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return &Conn{pool: pool, dialect: d.dialect, cfg: cfg}, nil
}

const defaultDatabase = "postgres"

func normalizeDatabase(database string) string {
	database = strings.TrimSpace(database)
	if database == "" {
		return defaultDatabase
	}
	return database
}

func buildDSN(cfg driver.ConnConfig) string {
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 5432
	}
	database := normalizeDatabase(cfg.Database)

	// postgres://user:password@host:port/dbname?sslmode=disable
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   host + ":" + strconv.Itoa(port),
		Path:   "/" + database,
	}

	q := u.Query()
	if cfg.SSL {
		q.Set("sslmode", "require")
	} else {
		q.Set("sslmode", "disable")
	}
	for k, v := range cfg.Params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// Conn PostgreSQL 连接实现
type Conn struct {
	pool    *pgxpool.Pool
	dialect driver.Dialect
	cfg     driver.ConnConfig
}

func (c *Conn) poolForDatabase(ctx context.Context, database string) (*pgxpool.Pool, func(), error) {
	target := normalizeDatabase(database)
	current := normalizeDatabase(c.cfg.Database)
	if target == current {
		return c.pool, func() {}, nil
	}

	cfg := c.cfg
	cfg.Database = target
	poolCfg, err := pgxpool.ParseConfig(buildDSN(cfg))
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: parse database %q config: %w", target, err)
	}
	poolCfg.MaxConns = 1
	poolCfg.MinConns = 1
	poolCfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: connect database %q: %w", target, err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("postgres: ping database %q: %w", target, err)
	}
	return pool, pool.Close, nil
}

// Ping 测试连接
func (c *Conn) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// WithDatabase PostgreSQL/Hologres 不能在连接内 USE database，需要连接到目标 database。
func (c *Conn) WithDatabase(ctx context.Context, database string) (driver.Conn, func(), error) {
	pool, cleanup, err := c.poolForDatabase(ctx, database)
	if err != nil {
		return nil, nil, err
	}
	cfg := c.cfg
	cfg.Database = normalizeDatabase(database)
	return &Conn{pool: pool, dialect: c.dialect, cfg: cfg}, cleanup, nil
}

// Close 关闭连接
func (c *Conn) Close() error {
	c.pool.Close()
	return nil
}

// Inspector 返回元数据查询器
func (c *Conn) Inspector() driver.Inspector {
	return newInspector(c)
}

// DDL 返回 DDL 生成器
func (c *Conn) DDL() driver.DDLGenerator {
	return newDDLGenerator(c)
}

// ServerVersion 返回 PG 版本
func (c *Conn) ServerVersion(ctx context.Context) (string, error) {
	var v string
	if err := c.pool.QueryRow(ctx, "SELECT version()").Scan(&v); err != nil {
		return "", err
	}
	return v, nil
}

// Execute 执行 SQL
func (c *Conn) Execute(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	if isQuery(sqlText) {
		return c.executeQuery(ctx, sqlText, args...)
	}
	return c.executeNonQuery(ctx, sqlText, args...)
}

func (c *Conn) executeQuery(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	rows, err := c.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: query: %w", err)
	}
	defer rows.Close()

	cols := buildColumnsFromPG(rows)
	data, err := scanAllPG(rows, len(cols))
	if err != nil {
		return nil, err
	}

	return &driver.Result{
		Columns: cols,
		Rows:    data,
	}, nil
}

func (c *Conn) executeNonQuery(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	tag, err := c.pool.Exec(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: exec: %w", err)
	}
	return &driver.Result{
		RowsAffected: tag.RowsAffected(),
	}, nil
}

// Stream 流式执行
func (c *Conn) Stream(ctx context.Context, sqlText string, args ...any) (driver.RowStream, error) {
	rows, err := c.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: stream: %w", err)
	}

	cols := buildColumnsFromPG(rows)
	return &pgRowStream{rows: rows, cols: cols}, nil
}

// pgRowStream 把 pgx.Rows 包装成 driver.RowStream
type pgRowStream struct {
	rows pgx.Rows
	cols []driver.Column
	err  error
}

func (s *pgRowStream) Columns() []driver.Column { return s.cols }
func (s *pgRowStream) Next() bool               { return s.rows.Next() }

func (s *pgRowStream) Scan() ([]any, error) {
	values, err := scanRowPG(s.rows)
	if err != nil {
		s.err = err
		return nil, err
	}
	return values, nil
}

func (s *pgRowStream) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.rows.Err()
}

func (s *pgRowStream) Close() error {
	s.rows.Close()
	return nil
}

// --- 辅助函数 ---

func isQuery(sqlText string) bool {
	for i := 0; i < len(sqlText); i++ {
		c := sqlText[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
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

func buildColumnsFromPG(rows pgx.Rows) []driver.Column {
	fields := rows.FieldDescriptions()
	cols := make([]driver.Column, len(fields))
	for i, f := range fields {
		cols[i] = driver.Column{
			Name:         string(f.Name),
			DatabaseType: oidTypeName(f.DataTypeOID),
			Nullable:     true,
		}
	}
	return cols
}

func scanRowPG(rows pgx.Rows) ([]any, error) {
	values, err := rows.Values()
	if err != nil {
		return nil, fmt.Errorf("postgres: scan: %w", err)
	}
	// 把 []byte 转 string；超大整数转 string 避免前端精度失真
	for i, v := range values {
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		values[i] = driver.NormalizeValue(v)
	}
	return values, nil
}
