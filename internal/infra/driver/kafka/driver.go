// Package kafka 提供 Kafka 驱动实现。
//
// Kafka 不是关系型数据库，但为了复用现有连接管理、对象树和结果面板，
// 这里实现 driver.Driver / driver.Conn 接口的最小子集：
//   - Inspector 列出 topics
//   - Execute 处理 DESCRIBE topic / LIST TOPICS 等元数据查询
//   - Stream 消费指定 topic 的消息
package kafka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cyancat/internal/infra/driver"

	"github.com/IBM/sarama"
)

// defaultConsumerGroup 默认消费者组（仅用于拉取消息）
const defaultConsumerGroup = "cyancat-consumer"

// Driver Kafka 驱动实现
type Driver struct{}

// New 创建 Kafka 驱动实例
func New() *Driver {
	return &Driver{}
}

// Type 返回驱动类型
func (d *Driver) Type() driver.DriverType {
	return driver.Kafka
}

// Dialect 返回 Kafka 方言（Kafka 没有 SQL 方言，复用占位实现）
func (d *Driver) Dialect() driver.Dialect {
	return &kafkaDialect{}
}

// Connect 建立 Kafka 连接
func (d *Driver) Connect(ctx context.Context, cfg driver.ConnConfig) (driver.Conn, error) {
	if cfg.Type != driver.Kafka {
		return nil, fmt.Errorf("kafka: expect type %q, got %q", driver.Kafka, cfg.Type)
	}

	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	brokers := parseBrokers(cfg.Host, cfg.Port)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka: at least one broker is required")
	}

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("kafka: new client: %w", err)
	}

	if err := client.RefreshMetadata(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("kafka: refresh metadata: %w", err)
	}

	return &Conn{
		client:   client,
		brokers:  brokers,
		config:   config,
		consumer: nil,
	}, nil
}

// Conn Kafka 连接实现
type Conn struct {
	client    sarama.Client
	brokers   []string
	config    *sarama.Config
	consumer  sarama.Consumer
	inspector driver.Inspector
}

// Ping 测试连接
func (c *Conn) Ping(ctx context.Context) error {
	if c.client == nil || c.client.Closed() {
		return fmt.Errorf("kafka: client is closed")
	}
	broker := c.client.LeastLoadedBroker()
	if broker == nil {
		return fmt.Errorf("kafka: no broker available")
	}
	connected, err := broker.Connected()
	if err != nil {
		return err
	}
	if !connected {
		return fmt.Errorf("kafka: broker not connected")
	}
	return nil
}

// WithDatabase Kafka 不需要 database 切换
func (c *Conn) WithDatabase(ctx context.Context, database string) (driver.Conn, func(), error) {
	return c, func() {}, nil
}

// Close 关闭连接
func (c *Conn) Close() error {
	if c.consumer != nil {
		_ = c.consumer.Close()
	}
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Inspector 返回 Kafka 元数据查询器
func (c *Conn) Inspector() driver.Inspector {
	if c.inspector == nil {
		c.inspector = newInspector(c)
	}
	return c.inspector
}

// DDL Kafka 不支持 DDL
func (c *Conn) DDL() driver.DDLGenerator {
	return &noopDDL{}
}

// ServerVersion 返回 Kafka 版本
func (c *Conn) ServerVersion(ctx context.Context) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("kafka: client not initialized")
	}
	return c.client.Config().Version.String(), nil
}

// Execute 执行类 SQL 元数据查询
//   - LIST TOPICS: 返回 topic 列表
//   - DESCRIBE topic_name: 返回 topic 分区信息
func (c *Conn) Execute(ctx context.Context, sqlText string, args ...any) (*driver.Result, error) {
	sqlText = strings.TrimSpace(sqlText)
	upper := strings.ToUpper(sqlText)

	if upper == "LIST TOPICS" {
		return c.listTopicsResult(ctx)
	}

	if strings.HasPrefix(upper, "DESCRIBE ") {
		topic := strings.TrimSpace(sqlText[len("DESCRIBE "):])
		return c.describeTopicResult(ctx, topic)
	}

	return nil, fmt.Errorf("kafka: unsupported query %q", sqlText)
}

// Stream 流式消费 topic 消息（返回最新 N 条）
// SQL 格式：CONSUME topic_name [LIMIT N]
func (c *Conn) Stream(ctx context.Context, sqlText string, args ...any) (driver.RowStream, error) {
	sqlText = strings.TrimSpace(sqlText)
	upper := strings.ToUpper(sqlText)

	if !strings.HasPrefix(upper, "CONSUME ") {
		return nil, fmt.Errorf("kafka: stream only supports CONSUME topic [LIMIT n]")
	}

	rest := strings.TrimSpace(sqlText[len("CONSUME "):])
	topic := rest
	limit := 100
	if idx := strings.LastIndex(strings.ToUpper(rest), " LIMIT "); idx >= 0 {
		topic = strings.TrimSpace(rest[:idx])
		fmt.Sscanf(strings.ToUpper(rest[idx:]), " LIMIT %d", &limit)
	}

	if topic == "" {
		return nil, fmt.Errorf("kafka: topic name is required")
	}

	return c.consumeTopic(ctx, topic, limit)
}

func (c *Conn) listTopicsResult(ctx context.Context) (*driver.Result, error) {
	topics, err := c.client.Topics()
	if err != nil {
		return nil, fmt.Errorf("kafka: list topics: %w", err)
	}

	rows := make([][]any, 0, len(topics))
	for _, t := range topics {
		rows = append(rows, []any{t})
	}

	return &driver.Result{
		Columns: []driver.Column{{Name: "topic"}},
		Rows:    driver.NormalizeRows(rows),
	}, nil
}

func (c *Conn) describeTopicResult(ctx context.Context, topic string) (*driver.Result, error) {
	partitions, err := c.client.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("kafka: describe topic %q: %w", topic, err)
	}

	rows := make([][]any, 0, len(partitions))
	for _, p := range partitions {
		rows = append(rows, []any{p})
	}

	return &driver.Result{
		Columns: []driver.Column{{Name: "partition"}},
		Rows:    driver.NormalizeRows(rows),
	}, nil
}

func (c *Conn) consumeTopic(ctx context.Context, topic string, limit int) (driver.RowStream, error) {
	partitions, err := c.client.Partitions(topic)
	if err != nil {
		return nil, fmt.Errorf("kafka: partitions for %q: %w", topic, err)
	}

	if c.consumer == nil {
		consumer, err := sarama.NewConsumerFromClient(c.client)
		if err != nil {
			return nil, fmt.Errorf("kafka: new consumer: %w", err)
		}
		c.consumer = consumer
	}

	// 简单策略：取每个分区的最新消息，直到凑够 limit
	msgs := make([]kafkaMessage, 0, limit)
	perPartition := limit
	if len(partitions) > 0 {
		perPartition = limit/len(partitions) + 1
	}

	for _, partition := range partitions {
		high, err := c.client.GetOffset(topic, partition, sarama.OffsetNewest)
		if err != nil {
			continue
		}
		start := high - int64(perPartition)
		if start < 0 {
			start = 0
		}

		pc, err := c.consumer.ConsumePartition(topic, partition, start)
		if err != nil {
			continue
		}

		count := 0
		timeout := time.After(2 * time.Second)
	loop:
		for count < perPartition {
			select {
			case msg := <-pc.Messages():
				if msg == nil {
					break loop
				}
				msgs = append(msgs, kafkaMessage{
					Partition: msg.Partition,
					Offset:    msg.Offset,
					Key:       string(msg.Key),
					Value:     string(msg.Value),
					Timestamp: msg.Timestamp.Format(time.RFC3339),
				})
				count++
			case <-timeout:
				break loop
			case <-ctx.Done():
				_ = pc.Close()
				return nil, ctx.Err()
			}
		}
		_ = pc.Close()
	}

	return &messageStream{msgs: msgs}, nil
}

type kafkaMessage struct {
	Partition int32
	Offset    int64
	Key       string
	Value     string
	Timestamp string
}

type messageStream struct {
	msgs  []kafkaMessage
	index int
	cols  []driver.Column
}

func (s *messageStream) Columns() []driver.Column {
	if s.cols == nil {
		s.cols = []driver.Column{
			{Name: "partition"},
			{Name: "offset"},
			{Name: "key"},
			{Name: "value"},
			{Name: "timestamp"},
		}
	}
	return s.cols
}

func (s *messageStream) Next() bool {
	if s.index >= len(s.msgs) {
		return false
	}
	s.index++
	return s.index <= len(s.msgs)
}

func (s *messageStream) Scan() ([]any, error) {
	if s.index <= 0 || s.index > len(s.msgs) {
		return nil, fmt.Errorf("kafka: scan out of range")
	}
	m := s.msgs[s.index-1]
	return []any{
		driver.NormalizeValue(m.Partition),
		driver.NormalizeValue(m.Offset),
		driver.NormalizeValue(m.Key),
		driver.NormalizeValue(m.Value),
		driver.NormalizeValue(m.Timestamp),
	}, nil
}

func (s *messageStream) Err() error {
	return nil
}

func (s *messageStream) Close() error {
	return nil
}

// parseBrokers 解析 broker 列表
func parseBrokers(host string, port int) []string {
	if host == "" {
		return nil
	}
	if port <= 0 {
		port = 9092
	}

	// 如果 host 已经是 host:port 列表
	parts := strings.Split(host, ",")
	for i, p := range parts {
		p = strings.TrimSpace(p)
		if !strings.Contains(p, ":") {
			p = fmt.Sprintf("%s:%d", p, port)
		}
		parts[i] = p
	}
	return parts
}

// kafkaDialect Kafka 占位方言
type kafkaDialect struct{}

func (d *kafkaDialect) QuoteIdent(ident string) string {
	return ident
}

func (d *kafkaDialect) Placeholder(n int) string {
	return "?"
}

func (d *kafkaDialect) DefaultLimit() int {
	return 100
}

// noopDDL Kafka 不支持 DDL
type noopDDL struct{}

func (n *noopDDL) CreateDatabase(spec driver.DatabaseSpec) (string, error) {
	return "", fmt.Errorf("kafka: CreateDatabase not supported")
}

func (n *noopDDL) DropDatabase(name string) (string, error) {
	return "", fmt.Errorf("kafka: DropDatabase not supported")
}

func (n *noopDDL) CreateTable(spec driver.TableSpec) (string, error) {
	return "", fmt.Errorf("kafka: CreateTable not supported")
}

func (n *noopDDL) AlterTable(spec driver.AlterTableSpec) (string, error) {
	return "", fmt.Errorf("kafka: AlterTable not supported")
}

func (n *noopDDL) DropTable(database, schema, name string) (string, error) {
	return "", fmt.Errorf("kafka: DropTable not supported")
}

func (n *noopDDL) RenameTable(database, schema, oldName, newName string) (string, error) {
	return "", fmt.Errorf("kafka: RenameTable not supported")
}

func (n *noopDDL) GetCreateTableDDL(ctx context.Context, database, schema, table string) (string, error) {
	return "", fmt.Errorf("kafka: GetCreateTableDDL not supported")
}
