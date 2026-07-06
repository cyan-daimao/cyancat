// Package redis 提供 Redis 数据源实现。
//
// Redis 不是关系型数据库，但为了复用现有连接管理、对象树和结果面板，
// 这里实现 driver.Driver / driver.Conn 接口的最小子集：
//   - Inspector 列出 databases（逻辑库）和 keys（作为"表"）
//   - Execute 处理 INFO / DBSIZE / TYPE / TTL / GET 等元数据查询
//   - Stream 执行 SCAN 或 LRANGE 等返回多条结果的操作
package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cyancat/internal/infra/driver"

	goredis "github.com/redis/go-redis/v9"
)

const defaultPort = 6379

// Driver Redis 驱动实现
type Driver struct{}

// New 创建 Redis 驱动实例
func New() *Driver {
	return &Driver{}
}

// Type 返回驱动类型
func (d *Driver) Type() driver.DriverType {
	return driver.Redis
}

// Dialect 返回 Redis 占位方言
func (d *Driver) Dialect() driver.Dialect {
	return &redisDialect{}
}

// Connect 建立 Redis 连接
func (d *Driver) Connect(ctx context.Context, cfg driver.ConnConfig) (driver.Conn, error) {
	if cfg.Type != driver.Redis {
		return nil, fmt.Errorf("redis: expect type %q, got %q", driver.Redis, cfg.Type)
	}

	port := cfg.Port
	if port <= 0 {
		port = defaultPort
	}
	host := cfg.Host
	if host == "" {
		host = "127.0.0.1"
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	opts := &goredis.Options{
		Addr:        addr,
		Username:    cfg.User,
		Password:    cfg.Password,
		DB:          parseDatabase(cfg.Database),
		DialTimeout: cfg.ConnectTimeout,
	}
	if opts.DialTimeout <= 0 {
		opts.DialTimeout = 5 * time.Second
	}

	client := goredis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	return &Conn{
		client: client,
		addr:   addr,
	}, nil
}

func parseDatabase(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// Conn Redis 连接实现
type Conn struct {
	client    *goredis.Client
	addr      string
	inspector driver.Inspector
}

// Ping 测试连接
func (c *Conn) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// WithDatabase Redis 通过 SELECT 切换逻辑库，返回当前连接即可
func (c *Conn) WithDatabase(ctx context.Context, database string) (driver.Conn, func(), error) {
	_ = database
	return c, func() {}, nil
}

// Close 关闭连接
func (c *Conn) Close() error {
	return c.client.Close()
}

// Inspector 返回 Redis 元数据查询器
func (c *Conn) Inspector() driver.Inspector {
	if c.inspector == nil {
		c.inspector = newInspector(c)
	}
	return c.inspector
}

// DDL Redis 不支持 DDL
func (c *Conn) DDL() driver.DDLGenerator {
	return &noopDDL{}
}

// ServerVersion 返回 Redis 服务端版本
func (c *Conn) ServerVersion(ctx context.Context) (string, error) {
	info, err := c.client.Info(ctx, "server").Result()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(info, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "redis_version:") {
			return strings.TrimPrefix(line, "redis_version:"), nil
		}
	}
	return "unknown", nil
}

// Execute 执行类 SQL 元数据查询：
//   - INFO [section]                返回服务器信息（单列多行文本）
//   - DB_SIZE / DBSIZE              返回当前库 key 数量
//   - TYPE key                      返回 key 类型
//   - TTL key                       返回 key 剩余秒数
//   - GET key / HGET key field      返回单个值
//   - HGETALL key                   返回 hash 全部字段（多行）
//   - LRANGE key 0 -1               返回列表全部元素
//   - SMEMBERS key                  返回集合成员
//   - ZRANGE key 0 -1               返回有序集合成员
func (c *Conn) Execute(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	sqlText = strings.TrimSpace(sqlText)
	tokens := tokenize(sqlText)

	if len(tokens) == 0 {
		return nil, fmt.Errorf("redis: empty command")
	}

	switch tokens[0] {
	case "INFO":
		section := ""
		if len(tokens) > 1 {
			section = tokens[1]
		}
		return c.executeInfo(ctx, section)
	case "DBSIZE", "DB_SIZE":
		return c.executeDBSize(ctx)
	case "TYPE":
		if len(tokens) < 2 {
			return nil, fmt.Errorf("redis: TYPE requires key")
		}
		return c.executeType(ctx, tokens[1])
	case "TTL":
		if len(tokens) < 2 {
			return nil, fmt.Errorf("redis: TTL requires key")
		}
		return c.executeTTL(ctx, tokens[1])
	case "GET":
		if len(tokens) < 2 {
			return nil, fmt.Errorf("redis: GET requires key")
		}
		return c.executeGet(ctx, tokens[1])
	case "HGET":
		if len(tokens) < 3 {
			return nil, fmt.Errorf("redis: HGET requires key and field")
		}
		return c.executeHGet(ctx, tokens[1], tokens[2])
	case "HGETALL":
		if len(tokens) < 2 {
			return nil, fmt.Errorf("redis: HGETALL requires key")
		}
		return c.executeHGetAll(ctx, tokens[1])
	case "LRANGE":
		if len(tokens) < 4 {
			return nil, fmt.Errorf("redis: LRANGE requires key start stop")
		}
		start, err1 := strconv.ParseInt(tokens[2], 10, 64)
		stop, err2 := strconv.ParseInt(tokens[3], 10, 64)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("redis: LRANGE start/stop must be integers")
		}
		return c.executeLRange(ctx, tokens[1], start, stop)
	case "SMEMBERS":
		if len(tokens) < 2 {
			return nil, fmt.Errorf("redis: SMEMBERS requires key")
		}
		return c.executeSMembers(ctx, tokens[1])
	case "ZRANGE":
		if len(tokens) < 4 {
			return nil, fmt.Errorf("redis: ZRANGE requires key start stop")
		}
		start, err1 := strconv.ParseInt(tokens[2], 10, 64)
		stop, err2 := strconv.ParseInt(tokens[3], 10, 64)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("redis: ZRANGE start/stop must be integers")
		}
		return c.executeZRange(ctx, tokens[1], start, stop)
	}

	return nil, fmt.Errorf("redis: unsupported command %q", sqlText)
}

// Stream 流式执行 SCAN 类查询：
//   - SCAN [cursor] [MATCH pattern] [COUNT n]
func (c *Conn) Stream(ctx context.Context, sqlText string, args ...any) (driver.RowStream, error) {
	sqlText = strings.TrimSpace(sqlText)
	tokens := tokenize(sqlText)

	if len(tokens) == 0 {
		return nil, fmt.Errorf("redis: empty command")
	}

	switch tokens[0] {
	case "SCAN":
		return c.streamScan(ctx, tokens)
	case "KEYS":
		if len(tokens) < 2 {
			return nil, fmt.Errorf("redis: KEYS requires pattern")
		}
		return c.streamKeys(ctx, tokens[1])
	}

	return nil, fmt.Errorf("redis: stream only supports SCAN / KEYS")
}

func tokenize(s string) []string {
	var tokens []string
	for _, t := range strings.Fields(s) {
		tokens = append(tokens, strings.ToUpper(t))
	}
	return tokens
}

func (c *Conn) executeInfo(ctx context.Context, section string) (*driver.Result, error) {
	var res string
	var err error
	if section == "" {
		res, err = c.client.Info(ctx).Result()
	} else {
		res, err = c.client.Info(ctx, section).Result()
	}
	if err != nil {
		return nil, err
	}
	rows := make([][]any, 0)
	for _, line := range strings.Split(res, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rows = append(rows, []any{line})
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "info"}},
		Rows:    driver.NormalizeRows(rows),
	}, nil
}

func (c *Conn) executeDBSize(ctx context.Context) (*driver.Result, error) {
	n, err := c.client.DBSize(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "db_size"}},
		Rows:    driver.NormalizeRows([][]any{{n}}),
	}, nil
}

func (c *Conn) executeType(ctx context.Context, key string) (*driver.Result, error) {
	t, err := c.client.Type(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "type"}},
		Rows:    driver.NormalizeRows([][]any{{t}}),
	}, nil
}

func (c *Conn) executeTTL(ctx context.Context, key string) (*driver.Result, error) {
	n, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "ttl_seconds"}},
		Rows:    driver.NormalizeRows([][]any{{int64(n / time.Second)}}),
	}, nil
}

func (c *Conn) executeGet(ctx context.Context, key string) (*driver.Result, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return &driver.Result{
			Columns: []driver.Column{{Name: "value"}},
			Rows:    driver.NormalizeRows([][]any{{"(nil)"}}),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "value"}},
		Rows:    driver.NormalizeRows([][]any{{val}}),
	}, nil
}

func (c *Conn) executeHGet(ctx context.Context, key, field string) (*driver.Result, error) {
	val, err := c.client.HGet(ctx, key, field).Result()
	if err == goredis.Nil {
		return &driver.Result{
			Columns: []driver.Column{{Name: "value"}},
			Rows:    driver.NormalizeRows([][]any{{"(nil)"}}),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "value"}},
		Rows:    driver.NormalizeRows([][]any{{val}}),
	}, nil
}

func (c *Conn) executeHGetAll(ctx context.Context, key string) (*driver.Result, error) {
	m, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	rows := make([][]any, 0, len(m))
	for k, v := range m {
		rows = append(rows, []any{k, v})
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "field"}, {Name: "value"}},
		Rows:    driver.NormalizeRows(rows),
	}, nil
}

func (c *Conn) executeLRange(ctx context.Context, key string, start, stop int64) (*driver.Result, error) {
	vals, err := c.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	rows := make([][]any, 0, len(vals))
	for i, v := range vals {
		rows = append(rows, []any{start + int64(i), v})
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "index"}, {Name: "value"}},
		Rows:    driver.NormalizeRows(rows),
	}, nil
}

func (c *Conn) executeSMembers(ctx context.Context, key string) (*driver.Result, error) {
	vals, err := c.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	rows := make([][]any, 0, len(vals))
	for _, v := range vals {
		rows = append(rows, []any{v})
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "member"}},
		Rows:    driver.NormalizeRows(rows),
	}, nil
}

func (c *Conn) executeZRange(ctx context.Context, key string, start, stop int64) (*driver.Result, error) {
	vals, err := c.client.ZRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}
	rows := make([][]any, 0, len(vals))
	for i, v := range vals {
		rows = append(rows, []any{start + int64(i), v})
	}
	return &driver.Result{
		Columns: []driver.Column{{Name: "rank"}, {Name: "member"}},
		Rows:    driver.NormalizeRows(rows),
	}, nil
}

func (c *Conn) streamScan(ctx context.Context, tokens []string) (driver.RowStream, error) {
	cursor := uint64(0)
	match := "*"
	count := int64(100)
	for i := 1; i < len(tokens); i += 2 {
		if i+1 >= len(tokens) {
			break
		}
		switch tokens[i] {
		case "MATCH":
			match = tokens[i+1]
		case "COUNT":
			count, _ = strconv.ParseInt(tokens[i+1], 10, 64)
		case "CURSOR":
			cursor, _ = strconv.ParseUint(tokens[i+1], 10, 64)
		}
	}
	if count <= 0 {
		count = 100
	}

	keys, nextCursor, err := c.client.Scan(ctx, cursor, match, count).Result()
	if err != nil {
		return nil, err
	}
	rows := make([][]any, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []any{k})
	}
	return &redisStream{
		cols: []driver.Column{{Name: "key"}, {Name: "next_cursor"}},
		rows: rows,
		extra: func(s *redisStream) {
			if len(s.rows) > 0 {
				s.rows[0] = append(s.rows[0], driver.NormalizeValue(nextCursor))
			}
		},
	}, nil
}

func (c *Conn) streamKeys(ctx context.Context, pattern string) (driver.RowStream, error) {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	rows := make([][]any, 0, len(keys))
	for _, k := range keys {
		rows = append(rows, []any{k})
	}
	return &redisStream{
		cols: []driver.Column{{Name: "key"}},
		rows: rows,
	}, nil
}

type redisStream struct {
	cols  []driver.Column
	rows  [][]any
	index int
	extra func(*redisStream)
}

func (s *redisStream) Columns() []driver.Column { return s.cols }
func (s *redisStream) Next() bool {
	if s.index >= len(s.rows) {
		return false
	}
	s.index++
	return s.index <= len(s.rows)
}
func (s *redisStream) Scan() ([]any, error) {
	if s.index <= 0 || s.index > len(s.rows) {
		return nil, fmt.Errorf("redis: scan out of range")
	}
	if s.extra != nil {
		s.extra(s)
		s.extra = nil
	}
	return s.rows[s.index-1], nil
}
func (s *redisStream) Err() error { return nil }
func (s *redisStream) Close() error { return nil }

// redisDialect Redis 占位方言
type redisDialect struct{}

func (d *redisDialect) QuoteIdent(ident string) string { return ident }
func (d *redisDialect) Placeholder(n int) string       { return "?" }
func (d *redisDialect) DefaultLimit() int              { return 100 }

// noopDDL Redis 不支持 DDL
type noopDDL struct{}

func (n *noopDDL) CreateDatabase(spec driver.DatabaseSpec) (string, error) {
	return "", fmt.Errorf("redis: CreateDatabase not supported")
}
func (n *noopDDL) DropDatabase(name string) (string, error) {
	return "", fmt.Errorf("redis: DropDatabase not supported")
}
func (n *noopDDL) CreateTable(spec driver.TableSpec) (string, error) {
	return "", fmt.Errorf("redis: CreateTable not supported")
}
func (n *noopDDL) AlterTable(spec driver.AlterTableSpec) (string, error) {
	return "", fmt.Errorf("redis: AlterTable not supported")
}
func (n *noopDDL) DropTable(database, schema, name string) (string, error) {
	return "", fmt.Errorf("redis: DropTable not supported")
}
func (n *noopDDL) RenameTable(database, schema, oldName, newName string) (string, error) {
	return "", fmt.Errorf("redis: RenameTable not supported")
}
func (n *noopDDL) GetCreateTableDDL(ctx context.Context, database, schema, table string) (string, error) {
	return "", fmt.Errorf("redis: GetCreateTableDDL not supported")
}
