// Package sqlite 提供 SQLite 文件数据库驱动实现
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cyancat/internal/infra/driver"

	_ "github.com/mattn/go-sqlite3"
)

// Driver SQLite 驱动实现
type Driver struct {
	dialect driver.Dialect
}

func New() *Driver {
	return &Driver{dialect: &sqliteDialect{}}
}

func (d *Driver) Type() driver.DriverType {
	return driver.SQLite
}

func (d *Driver) Dialect() driver.Dialect {
	return d.dialect
}

func (d *Driver) Connect(ctx context.Context, cfg driver.ConnConfig) (driver.Conn, error) {
	if cfg.Type != driver.SQLite {
		return nil, fmt.Errorf("sqlite: expect type %q, got %q", driver.SQLite, cfg.Type)
	}

	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
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
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}
	if _, err := db.ExecContext(connectCtx, "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: enable foreign keys: %w", err)
	}

	conn := &Conn{db: db}
	conn.inspector = newInspector(conn)
	conn.ddl = newDDLGenerator(conn)
	return conn, nil
}

func buildDSN(cfg driver.ConnConfig) (string, error) {
	path := strings.TrimSpace(cfg.Host)
	if path == "" {
		return "", fmt.Errorf("sqlite: database file is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("sqlite: resolve database file: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("sqlite: database file does not exist: %s", abs)
		}
		return "", fmt.Errorf("sqlite: stat database file: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("sqlite: database path is a directory: %s", abs)
	}

	u := url.URL{Scheme: "file", Path: abs}
	q := u.Query()
	q.Set("mode", "rw")
	q.Set("_foreign_keys", "on")
	q.Set("_busy_timeout", "5000")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// Conn SQLite 连接实现
type Conn struct {
	db        *sql.DB
	inspector driver.Inspector
	ddl       driver.DDLGenerator
}

func (c *Conn) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *Conn) WithDatabase(ctx context.Context, database string) (driver.Conn, func(), error) {
	_ = ctx
	_ = database
	return c, func() {}, nil
}

func (c *Conn) Execute(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	if isQuery(sqlText) {
		return c.executeQuery(ctx, sqlText, args...)
	}
	return c.executeNonQuery(ctx, sqlText, args...)
}

func (c *Conn) executeQuery(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	rows, err := c.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlite: query: %w", err)
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
	return &driver.Result{Columns: cols, Rows: data}, nil
}

func (c *Conn) executeNonQuery(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	res, err := c.db.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlite: exec: %w", err)
	}
	affected, _ := res.RowsAffected()
	lastID, _ := res.LastInsertId()
	return &driver.Result{RowsAffected: affected, LastInsertID: lastID}, nil
}

func (c *Conn) Stream(ctx context.Context, sqlText string, args ...any) (driver.RowStream, error) {
	rows, err := c.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlite: stream: %w", err)
	}
	cols, err := buildColumns(rows)
	if err != nil {
		_ = rows.Close()
		return nil, err
	}
	return &rowStream{rows: rows, cols: cols}, nil
}

func (c *Conn) Inspector() driver.Inspector {
	return c.inspector
}

func (c *Conn) DDL() driver.DDLGenerator {
	return c.ddl
}

func (c *Conn) ServerVersion(ctx context.Context) (string, error) {
	var v string
	if err := c.db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&v); err != nil {
		return "", err
	}
	return v, nil
}

func (c *Conn) Close() error {
	return c.db.Close()
}

type rowStream struct {
	rows *sql.Rows
	cols []driver.Column
	err  error
}

func (s *rowStream) Columns() []driver.Column {
	return s.cols
}

func (s *rowStream) Next() bool {
	return s.rows.Next()
}

func (s *rowStream) Scan() ([]any, error) {
	values, err := scanRow(s.rows, len(s.cols))
	if err != nil {
		s.err = err
		return nil, err
	}
	return values, nil
}

func (s *rowStream) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.rows.Err()
}

func (s *rowStream) Close() error {
	return s.rows.Close()
}

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
			startsWithFold(prefix, "PRAGMA"),
			startsWithFold(prefix, "EXPLAI"),
			startsWithFold(prefix, "WITH"):
			return true
		default:
			return false
		}
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

func buildColumns(rows *sql.Rows) ([]driver.Column, error) {
	names, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("sqlite: columns: %w", err)
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("sqlite: column types: %w", err)
	}
	cols := make([]driver.Column, len(names))
	for i, name := range names {
		col := driver.Column{Name: name, Nullable: true}
		if i < len(types) {
			col.DatabaseType = types[i].DatabaseTypeName()
			if nullable, ok := types[i].Nullable(); ok {
				col.Nullable = nullable
			}
		}
		cols[i] = col
	}
	return cols, nil
}

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

func scanRow(rows *sql.Rows, colCount int) ([]any, error) {
	values := make([]any, colCount)
	scanPtrs := make([]any, colCount)
	for i := range values {
		scanPtrs[i] = &values[i]
	}
	if err := rows.Scan(scanPtrs...); err != nil {
		return nil, fmt.Errorf("sqlite: scan: %w", err)
	}
	for i, v := range values {
		if b, ok := v.([]byte); ok {
			values[i] = string(b)
		}
	}
	return values, nil
}
