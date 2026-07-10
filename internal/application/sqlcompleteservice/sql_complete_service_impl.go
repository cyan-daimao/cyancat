package sqlcompleteservice

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cyancat/internal/application/schemaservice"
	"cyancat/internal/application/sqlcompleteservice/parser"
)

// ServiceImpl SQL 补全服务实现。
type ServiceImpl struct {
	schemaSvc schemaservice.SchemaService
	parsers   map[string]parser.Parser

	// columnCache 字段缓存：key = connID:database:schema:table
	columnCache map[string]columnCacheEntry
	cacheMu     sync.RWMutex
	cacheMaxAge time.Duration
}

type columnCacheEntry struct {
	columns   []schemaservice.ColumnBO
	cachedAt  time.Time
}

// NewServiceImpl 创建补全服务。
func NewServiceImpl(schemaSvc schemaservice.SchemaService) *ServiceImpl {
	svc := &ServiceImpl{
		schemaSvc:   schemaSvc,
		parsers:     make(map[string]parser.Parser),
		columnCache: make(map[string]columnCacheEntry),
		cacheMaxAge: 5 * time.Minute,
	}
	go svc.cacheCleaner()
	return svc
}

// Complete 实现 Service 接口。
func (s *ServiceImpl) Complete(query *CompleteQuery) (*CompleteResult, error) {
	if query == nil {
		return nil, fmt.Errorf("sqlcompleteservice: query cannot be nil")
	}
	if query.ConnID <= 0 {
		return nil, fmt.Errorf("sqlcompleteservice: connID must be positive")
	}

	p := s.getParser(query.ConnectionType)
	parseRes, err := p.Parse(query.SQL, query.CursorLine, query.CursorColumn)
	if err != nil {
		// 解析失败时尽量 fallback 到关键字补全
		parseRes = &parser.ParseResult{
			Context: parser.CtxKeyword,
			Tables:  []parser.TableRef{},
			CTEs:    []string{},
		}
	}

	ctx := context.Background()
	candidates := make([]CompleteCandidate, 0)

	// Parser 可能把不完整的 FROM/JOIN 解析成无表的 SELECT，此时根据前缀重新判断上下文
	if parseRes.Context == parser.CtxColumn && len(parseRes.Tables) == 0 && !parseRes.IsMemberAccess {
		if isTableContextPrefix(query.SQL, query.CursorLine, query.CursorColumn) {
			parseRes.Context = parser.CtxTable
		}
	}

	switch parseRes.Context {
	case parser.CtxColumn:
		candidates = s.completeColumn(ctx, query, parseRes)
	case parser.CtxTable:
		candidates = s.completeTable(ctx, query)
	case parser.CtxKeyword:
		candidates = s.completeKeyword()
	default:
		// 默认同时给出关键字和相关字段/表
		candidates = append(candidates, s.completeKeyword()...)
		candidates = append(candidates, s.completeColumn(ctx, query, parseRes)...)
	}

	return &CompleteResult{Candidates: candidates}, nil
}

func (s *ServiceImpl) getParser(connectionType string) parser.Parser {
	key := strings.ToLower(connectionType)
	if p, ok := s.parsers[key]; ok {
		return p
	}
	p := parser.New(connectionType)
	s.parsers[key] = p
	return p
}

func (s *ServiceImpl) completeColumn(ctx context.Context, query *CompleteQuery, parseRes *parser.ParseResult) []CompleteCandidate {
	var tables []parser.TableRef
	if parseRes.IsMemberAccess && parseRes.TablePrefix != "" {
		// alias. 或 table.：只补指定前缀的字段
		tables = s.resolvePrefixToTables(parseRes, query)
	} else {
		// 无前缀：补当前 SQL 涉及的所有表的字段
		tables = parseRes.Tables
	}

	candidates := make([]CompleteCandidate, 0)
	seen := make(map[string]bool)
	for _, t := range tables {
		columns, err := s.getColumns(ctx, query, t)
		if err != nil {
			continue
		}
		for i, col := range columns {
			label := col.Name
			insertText := s.quoteIdentifier(col.Name, query.ConnectionType)
			if !parseRes.IsMemberAccess && len(tables) > 1 {
				// 多表且无前缀时，用 table.column 形式避免歧义
				label = fmt.Sprintf("%s.%s", t.Name, col.Name)
				insertText = fmt.Sprintf("%s.%s", s.quoteIdentifier(t.Name, query.ConnectionType), s.quoteIdentifier(col.Name, query.ConnectionType))
			}
			if seen[label] {
				continue
			}
			seen[label] = true
			candidates = append(candidates, CompleteCandidate{
				Label:      label,
				Kind:       KindColumn,
				Detail:     col.DatabaseType,
				InsertText: insertText,
				SortText:   fmt.Sprintf("1_%04d_%s", i, col.Name),
			})
		}
	}
	return candidates
}

func (s *ServiceImpl) completeTable(ctx context.Context, query *CompleteQuery) []CompleteCandidate {
	// 表名补全：按当前输入前缀搜索；无前缀时返回前 50 个表/视图作为兜底
	prefix := strings.TrimSpace(query.Prefix)

	candidates := make([]CompleteCandidate, 0)
	var list []*schemaservice.TableBO
	var err error

	if prefix == "" {
		list, err = s.schemaSvc.ListTables(&schemaservice.ListTablesQuery{
			ConnID:   query.ConnID,
			Database: query.Database,
			Schema:   query.Schema,
			Limit:    50,
			Offset:   0,
		})
		if err != nil {
			return candidates
		}
		views, _ := s.schemaSvc.ListViews(&schemaservice.ListTablesQuery{
			ConnID:   query.ConnID,
			Database: query.Database,
			Schema:   query.Schema,
			Limit:    50,
			Offset:   0,
		})
		for _, v := range views {
			list = append(list, &schemaservice.TableBO{
				Name: v.Name,
				Type: "VIEW",
			})
		}
	} else {
		list, err = s.schemaSvc.SearchTables(&schemaservice.SearchTablesQuery{
			ConnID:   query.ConnID,
			Database: query.Database,
			Schema:   query.Schema,
			Keyword:  prefix,
			Limit:    50,
		})
		if err != nil {
			return candidates
		}
	}

	for i, t := range list {
		kind := KindTable
		if strings.EqualFold(t.Type, "VIEW") || strings.EqualFold(t.Type, "v") {
			kind = KindView
		}
		candidates = append(candidates, CompleteCandidate{
			Label:      t.Name,
			Kind:       kind,
			Detail:     t.Comment,
			InsertText: s.quoteIdentifier(t.Name, query.ConnectionType),
			SortText:   fmt.Sprintf("1_%04d_%s", i, t.Name),
		})
	}
	return candidates
}

func (s *ServiceImpl) completeKeyword() []CompleteCandidate {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "AND", "OR", "NOT", "IN", "LIKE", "IS", "NULL",
		"INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE",
		"CREATE", "TABLE", "DROP", "ALTER", "INDEX", "VIEW", "DATABASE",
		"JOIN", "INNER", "LEFT", "RIGHT", "OUTER", "ON", "AS",
		"GROUP", "BY", "ORDER", "HAVING", "LIMIT", "OFFSET", "DISTINCT",
		"UNION", "ALL", "CASE", "WHEN", "THEN", "ELSE", "END",
	}
	candidates := make([]CompleteCandidate, 0, len(keywords))
	for i, kw := range keywords {
		candidates = append(candidates, CompleteCandidate{
			Label:      kw,
			Kind:       KindKeyword,
			Detail:     "SQL keyword",
			InsertText: kw,
			SortText:   fmt.Sprintf("9_%04d_%s", i, kw),
		})
	}
	return candidates
}

// resolvePrefixToTables 把 alias/table 前缀解析成实际的 TableRef。
func (s *ServiceImpl) resolvePrefixToTables(parseRes *parser.ParseResult, query *CompleteQuery) []parser.TableRef {
	prefix := strings.TrimSpace(parseRes.TablePrefix)
	for _, t := range parseRes.Tables {
		if strings.EqualFold(t.Alias, prefix) {
			return []parser.TableRef{t}
		}
		if strings.EqualFold(t.Name, prefix) {
			return []parser.TableRef{t}
		}
		// PostgreSQL 可能带 schema 前缀：schema.table
		if t.Schema != "" && strings.EqualFold(t.Schema+"."+t.Name, prefix) {
			return []parser.TableRef{t}
		}
	}
	// 未找到时尝试用前缀作为表名直接查
	return []parser.TableRef{{Name: prefix, Schema: query.Schema}}
}

func (s *ServiceImpl) getColumns(ctx context.Context, query *CompleteQuery, ref parser.TableRef) ([]schemaservice.ColumnBO, error) {
	schema := ref.Schema
	if schema == "" {
		schema = query.Schema
	}
	name := ref.Name
	if name == "" {
		return nil, fmt.Errorf("empty table name")
	}

	cacheKey := fmt.Sprintf("%d:%s:%s:%s", query.ConnID, query.Database, schema, name)

	s.cacheMu.RLock()
	entry, ok := s.columnCache[cacheKey]
	s.cacheMu.RUnlock()
	if ok && time.Since(entry.cachedAt) < s.cacheMaxAge {
		return entry.columns, nil
	}

	detail, err := s.schemaSvc.DescribeTable(&schemaservice.DescribeTableQuery{
		ConnID:   query.ConnID,
		Database: query.Database,
		Schema:   schema,
		Table:    name,
	})
	if err != nil {
		return nil, err
	}

	s.cacheMu.Lock()
	s.columnCache[cacheKey] = columnCacheEntry{
		columns:  detail.Columns,
		cachedAt: time.Now(),
	}
	s.cacheMu.Unlock()
	return detail.Columns, nil
}

func (s *ServiceImpl) quoteIdentifier(name, connectionType string) string {
	// 简单名不需要引号；包含特殊字符或关键字时按需引号
	if name == "" {
		return name
	}
	// 保守策略：只包含字母数字下划线且不以下划线开头的不引号
	if isSimpleIdentifier(name) {
		return name
	}
	switch strings.ToLower(connectionType) {
	case "postgres", "postgresql":
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	default:
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	}
}

func isSimpleIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if r != '_' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
				return false
			}
		} else {
			if r != '_' && (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
				return false
			}
		}
	}
	return true
}

// cacheCleaner 定期清理过期缓存。
func (s *ServiceImpl) cacheCleaner() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.cacheMu.Lock()
		now := time.Now()
		for k, v := range s.columnCache {
			if now.Sub(v.cachedAt) > s.cacheMaxAge {
				delete(s.columnCache, k)
			}
		}
		s.cacheMu.Unlock()
	}
}

// ClearConnectionCache 连接关闭时清空该连接的缓存。
func (s *ServiceImpl) ClearConnectionCache(connID int64) {
	prefix := fmt.Sprintf("%d:", connID)
	s.cacheMu.Lock()
	for k := range s.columnCache {
		if strings.HasPrefix(k, prefix) {
			delete(s.columnCache, k)
		}
	}
	s.cacheMu.Unlock()
}

// isTableContextPrefix 判断光标前是否处于 FROM/JOIN 后等待表名的位置。
func isTableContextPrefix(sql string, cursorLine, cursorColumn int) bool {
	offset := cursorOffset(sql, cursorLine, cursorColumn)
	if offset < 0 || offset > len(sql) {
		offset = len(sql)
	}
	prefix := strings.TrimSpace(sql[:offset])
	upper := strings.ToUpper(prefix)
	if strings.HasSuffix(upper, " FROM") || strings.HasSuffix(upper, " JOIN") || strings.HasSuffix(upper, ",") {
		return true
	}
	return false
}

// cursorOffset 将 1-based 行列转换为字节偏移（与 parser 包保持一致）。
func cursorOffset(sql string, line, column int) int {
	if line <= 1 {
		if column <= 1 {
			return 0
		}
		return min(column-1, len(sql))
	}
	currentLine := 1
	for i, c := range []byte(sql) {
		if c == '\n' {
			currentLine++
			if currentLine == line {
				next := i + column
				if next > len(sql) {
					return len(sql)
				}
				return next
			}
		}
	}
	return len(sql)
}
