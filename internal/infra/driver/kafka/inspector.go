package kafka

import (
	"context"
	"fmt"

	"cyancat/internal/infra/driver"
)

// inspector Kafka 元数据查询器
type inspector struct {
	conn *Conn
}

func newInspector(c *Conn) driver.Inspector {
	return &inspector{conn: c}
}

// ListDatabases Kafka 把 "database" 概念映射为 topic 分类占位
func (i *inspector) ListDatabases(ctx context.Context) ([]driver.Database, error) {
	topics, err := i.conn.client.Topics()
	if err != nil {
		return nil, fmt.Errorf("kafka/inspector: list topics: %w", err)
	}
	result := make([]driver.Database, 0, len(topics))
	for _, t := range topics {
		result = append(result, driver.Database{Name: t})
	}
	return result, nil
}

// ListSchemas Kafka 没有 schema，返回单个占位
func (i *inspector) ListSchemas(ctx context.Context, database string) ([]driver.Schema, error) {
	return []driver.Schema{{Name: "default"}}, nil
}

// ListTables 列出 topic 下的分区（作为"表"）
func (i *inspector) ListTables(ctx context.Context, database, schema string) ([]driver.Table, error) {
	partitions, err := i.conn.client.Partitions(database)
	if err != nil {
		return nil, fmt.Errorf("kafka/inspector: list partitions for %q: %w", database, err)
	}
	result := make([]driver.Table, 0, len(partitions))
	for _, p := range partitions {
		result = append(result, driver.Table{
			Name: fmt.Sprintf("partition_%d", p),
			Type: "PARTITION",
		})
	}
	return result, nil
}

// ListViews Kafka 没有视图
func (i *inspector) ListViews(ctx context.Context, database, schema string) ([]driver.View, error) {
	return []driver.View{}, nil
}

// DescribeTable 描述 topic 元数据
func (i *inspector) DescribeTable(ctx context.Context, database, schema, table string) (*driver.TableDetail, error) {
	partitions, err := i.conn.client.Partitions(database)
	if err != nil {
		return nil, fmt.Errorf("kafka/inspector: partitions for %q: %w", database, err)
	}

	detail := &driver.TableDetail{
		Name:     database,
		Database: database,
		Schema:   schema,
		Comment:  "Kafka topic",
		Columns: []driver.Column{
			{Name: "partition", DatabaseType: "int32"},
			{Name: "offset", DatabaseType: "int64"},
			{Name: "key", DatabaseType: "string"},
			{Name: "value", DatabaseType: "string"},
			{Name: "timestamp", DatabaseType: "string"},
		},
	}

	for _, p := range partitions {
		detail.Indexes = append(detail.Indexes, driver.Index{
			Name:    fmt.Sprintf("partition_%d", p),
			Columns: []string{"partition"},
		})
	}

	return detail, nil
}

// ListIndexes 列出分区索引（占位）
func (i *inspector) ListIndexes(ctx context.Context, database, schema, table string) ([]driver.Index, error) {
	partitions, err := i.conn.client.Partitions(database)
	if err != nil {
		return nil, err
	}
	result := make([]driver.Index, 0, len(partitions))
	for _, p := range partitions {
		result = append(result, driver.Index{Name: fmt.Sprintf("partition_%d", p)})
	}
	return result, nil
}

// ListForeignKeys Kafka 没有外键
func (i *inspector) ListForeignKeys(ctx context.Context, database, schema, table string) ([]driver.ForeignKey, error) {
	return []driver.ForeignKey{}, nil
}

// ListCharsets Kafka 没有字符集
func (i *inspector) ListCharsets(ctx context.Context) ([]driver.Charset, error) {
	return []driver.Charset{}, nil
}

// ListCollations Kafka 没有排序规则
func (i *inspector) ListCollations(ctx context.Context, charset string) ([]driver.Collation, error) {
	return []driver.Collation{}, nil
}
