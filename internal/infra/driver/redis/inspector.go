package redis

import (
	"context"
	"fmt"
	"strconv"

	"cyancat/internal/infra/driver"
)

// inspector Redis 元数据查询器。
// 映射关系：
//   - ListDatabases  -> Redis 逻辑库（db0 ~ dbN），固定展示 db0~db15
//   - ListSchemas    -> 单个占位 schema
//   - ListTables     -> 当前逻辑库中的 keys（作为"表"）
//   - DescribeTable  -> key 的类型、大小、TTL 等元信息
type inspector struct {
	conn *Conn
}

func newInspector(c *Conn) driver.Inspector {
	return &inspector{conn: c}
}

// ListDatabases 列出 Redis 逻辑库（db0 ~ db15）
func (i *inspector) ListDatabases(ctx context.Context) ([]driver.Database, error) {
	result := make([]driver.Database, 0, 16)
	for n := 0; n < 16; n++ {
		result = append(result, driver.Database{Name: fmt.Sprintf("db%d", n)})
	}
	return result, nil
}

// ListSchemas Redis 没有 schema，返回单个占位
func (i *inspector) ListSchemas(ctx context.Context, database string) ([]driver.Schema, error) {
	return []driver.Schema{{Name: "default"}}, nil
}

// ListTables 列出指定 database 中的 keys（最多 500 个）
func (i *inspector) ListTables(ctx context.Context, database, schema string) ([]driver.Table, error) {
	_ = database
	keys, _, err := i.conn.client.Scan(ctx, 0, "*", 500).Result()
	if err != nil {
		return nil, fmt.Errorf("redis/inspector: scan keys: %w", err)
	}

	result := make([]driver.Table, 0, len(keys))
	for _, k := range keys {
		result = append(result, driver.Table{Name: k, Type: "KEY"})
	}
	return result, nil
}

// ListViews Redis 没有视图
func (i *inspector) ListViews(ctx context.Context, database, schema string) ([]driver.View, error) {
	return []driver.View{}, nil
}

// DescribeTable 描述 Redis key 的元信息
func (i *inspector) DescribeTable(ctx context.Context, database, schema, table string) (*driver.TableDetail, error) {
	key := table
	if key == "" {
		return nil, fmt.Errorf("redis/inspector: key is required")
	}

	typeVal, err := i.conn.client.Type(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis/inspector: type: %w", err)
	}

	ttl, _ := i.conn.client.TTL(ctx, key).Result()
	var size int64
	switch typeVal {
	case "string":
		size, _ = i.conn.client.StrLen(ctx, key).Result()
	case "list":
		size, _ = i.conn.client.LLen(ctx, key).Result()
	case "set":
		size, _ = i.conn.client.SCard(ctx, key).Result()
	case "zset":
		size, _ = i.conn.client.ZCard(ctx, key).Result()
	case "hash":
		size, _ = i.conn.client.HLen(ctx, key).Result()
	}

	detail := &driver.TableDetail{
		Name:     key,
		Database: database,
		Schema:   schema,
		Comment:  fmt.Sprintf("Redis %s key", typeVal),
		Columns: []driver.Column{
			{Name: "attribute", DatabaseType: "string"},
			{Name: "value", DatabaseType: "string"},
		},
	}

	rows := [][]string{
		{"type", typeVal},
		{"ttl_seconds", strconv.FormatInt(int64(ttl.Seconds()), 10)},
		{"size", strconv.FormatInt(size, 10)},
	}
	for _, r := range rows {
		detail.Indexes = append(detail.Indexes, driver.Index{
			Name:    r[0],
			Columns: []string{r[1]},
		})
	}

	return detail, nil
}

// ListIndexes 列出 key 的元信息索引（占位）
func (i *inspector) ListIndexes(ctx context.Context, database, schema, table string) ([]driver.Index, error) {
	detail, err := i.DescribeTable(ctx, database, schema, table)
	if err != nil {
		return nil, err
	}
	return detail.Indexes, nil
}

// ListForeignKeys Redis 没有外键
func (i *inspector) ListForeignKeys(ctx context.Context, database, schema, table string) ([]driver.ForeignKey, error) {
	return []driver.ForeignKey{}, nil
}

// ListCharsets Redis 没有字符集
func (i *inspector) ListCharsets(ctx context.Context) ([]driver.Charset, error) {
	return []driver.Charset{}, nil
}

// ListCollations Redis 没有排序规则
func (i *inspector) ListCollations(ctx context.Context, charset string) ([]driver.Collation, error) {
	return []driver.Collation{}, nil
}
