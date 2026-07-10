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
	// tableCache 表名缓存：key = connID:database:schema:keyword
	// keyword 为空字符串表示“前 N 个兜底表/视图”，不同数据源/schema 自然隔离
	tableCache map[string]tableCacheEntry
	// schemaCache 模式缓存：key = connID:database
	schemaCache map[string]schemaCacheEntry
	cacheMu     sync.RWMutex
	cacheMaxAge time.Duration
}

type columnCacheEntry struct {
	columns  []schemaservice.ColumnBO
	cachedAt time.Time
}

type tableCacheEntry struct {
	tables   []*schemaservice.TableBO
	cachedAt time.Time
}

type schemaCacheEntry struct {
	schemas  []string
	cachedAt time.Time
}

// NewServiceImpl 创建补全服务。
func NewServiceImpl(schemaSvc schemaservice.SchemaService) *ServiceImpl {
	svc := &ServiceImpl{
		schemaSvc:   schemaSvc,
		parsers:     make(map[string]parser.Parser),
		columnCache: make(map[string]columnCacheEntry),
		tableCache:  make(map[string]tableCacheEntry),
		schemaCache: make(map[string]schemaCacheEntry),
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
		// PostgreSQL 中 schema. 应优先提示该 schema 下的表，而不是字段
		if isPostgresLike(query.ConnectionType) && parseRes.IsMemberAccess && parseRes.TablePrefix != "" {
			schemaCandidates := s.completeSchemaTable(ctx, query, parseRes.TablePrefix, "")
			if len(schemaCandidates) > 0 {
				candidates = schemaCandidates
				break
			}
		}
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

	// PostgreSQL 支持 schema.table 和 schema 名提示
	if isPostgresLike(query.ConnectionType) {
		if dotIdx := strings.Index(prefix, "."); dotIdx >= 0 {
			schemaPart := strings.TrimSpace(prefix[:dotIdx])
			tablePart := strings.TrimSpace(prefix[dotIdx+1:])
			schemaCandidates := s.completeSchemaTable(ctx, query, schemaPart, tablePart)
			if len(schemaCandidates) > 0 {
				return schemaCandidates
			}
		}

		if prefix == "" {
			// 在数据库层（未指定 schema）时，空前缀优先提示 schema 列表
			if query.Schema == "" {
				return s.schemaCandidates(ctx, query, "", 100)
			}
			// 在 schema 层时保持原行为：提示该 schema 下的表
			return s.completeTableRaw(ctx, query, prefix)
		}

		// 非空前缀：同时提示匹配的 schema 和表
		candidates := s.schemaCandidates(ctx, query, prefix, 50)
		candidates = append(candidates, s.completeTableRaw(ctx, query, prefix)...)
		return candidates
	}

	return s.completeTableRaw(ctx, query, prefix)
}

// schemaCandidates 返回当前数据库下的 schema 候选（PostgreSQL 专用）。
// prefix 为空时返回前 limit 个 schema；非空时按模糊匹配过滤。
func (s *ServiceImpl) schemaCandidates(ctx context.Context, query *CompleteQuery, prefix string, limit int) []CompleteCandidate {
	if query.Database == "" {
		return nil
	}
	schemas, err := s.getSchemas(ctx, query)
	if err != nil {
		return nil
	}
	lowerPrefix := strings.ToLower(prefix)
	candidates := make([]CompleteCandidate, 0)
	idx := 0
	for _, sc := range schemas {
		if prefix != "" && !strings.Contains(strings.ToLower(sc), lowerPrefix) {
			continue
		}
		candidates = append(candidates, CompleteCandidate{
			Label:      sc,
			Kind:       KindSchema,
			Detail:     "schema",
			InsertText: s.quoteIdentifier(sc, query.ConnectionType),
			SortText:   fmt.Sprintf("0_%04d_%s", idx, sc),
		})
		idx++
		if prefix == "" && idx >= limit {
			break
		}
	}
	return candidates
}

// completeTableRaw 执行实际的表/视图搜索并缓存，不附加 schema 候选。
func (s *ServiceImpl) completeTableRaw(ctx context.Context, query *CompleteQuery, prefix string) []CompleteCandidate {
	candidates := make([]CompleteCandidate, 0)

	cacheKey := s.tableCacheKey(query, prefix)
	s.cacheMu.RLock()
	entry, ok := s.tableCache[cacheKey]
	s.cacheMu.RUnlock()
	if ok && time.Since(entry.cachedAt) < s.cacheMaxAge {
		return s.tablesToCandidates(entry.tables, query.ConnectionType)
	}

	// PostgreSQL 非空前缀时，优先从当前 schema 的全量缓存中模糊过滤，避免被 SearchTables LIMIT 截断
	if isPostgresLike(query.ConnectionType) && prefix != "" {
		emptyKey := fmt.Sprintf("%d:%s:%s:", query.ConnID, query.Database, query.Schema)
		s.cacheMu.RLock()
		entry, ok = s.tableCache[emptyKey]
		s.cacheMu.RUnlock()
		if ok && time.Since(entry.cachedAt) < s.cacheMaxAge {
			filtered := fuzzyFilterTables(entry.tables, prefix)
			return s.tablesToCandidates(filtered, query.ConnectionType)
		}
	}

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

	s.cacheMu.Lock()
	s.tableCache[cacheKey] = tableCacheEntry{
		tables:   list,
		cachedAt: time.Now(),
	}
	s.cacheMu.Unlock()

	return s.tablesToCandidates(list, query.ConnectionType)
}

func (s *ServiceImpl) tablesToCandidates(list []*schemaservice.TableBO, connectionType string) []CompleteCandidate {
	candidates := make([]CompleteCandidate, 0, len(list))
	for i, t := range list {
		kind := KindTable
		if strings.EqualFold(t.Type, "VIEW") || strings.EqualFold(t.Type, "v") {
			kind = KindView
		}
		candidates = append(candidates, CompleteCandidate{
			Label:      t.Name,
			Kind:       kind,
			Detail:     t.Comment,
			InsertText: s.quoteIdentifier(t.Name, connectionType),
			SortText:   fmt.Sprintf("1_%04d_%s", i, t.Name),
		})
	}
	return candidates
}

func (s *ServiceImpl) tableCacheKey(query *CompleteQuery, prefix string) string {
	return fmt.Sprintf("%d:%s:%s:%s", query.ConnID, query.Database, query.Schema, strings.ToLower(prefix))
}

func isPostgresLike(connectionType string) bool {
	t := strings.ToLower(connectionType)
	return t == "postgres" || t == "postgresql"
}

// completeSchemaTable 提示指定 schema 下的表（PostgreSQL 专用）。
// schemaPrefix 为 schema 名，tablePrefix 为可选的表名前缀（模糊搜索用）。
func (s *ServiceImpl) completeSchemaTable(ctx context.Context, query *CompleteQuery, schemaPrefix, tablePrefix string) []CompleteCandidate {
	schemaPrefix = strings.TrimSpace(schemaPrefix)
	tablePrefix = strings.TrimSpace(tablePrefix)
	if schemaPrefix == "" {
		return nil
	}

	// 先确认该 prefix 确实是一个存在的 schema
	schemas, err := s.getSchemas(ctx, query)
	if err != nil {
		return nil
	}
	found := false
	for _, sc := range schemas {
		if strings.EqualFold(sc, schemaPrefix) {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	// 优先从缓存中全量模糊过滤，避免被 SearchTables 的 LIMIT 截断
	cacheKey := fmt.Sprintf("%d:%s:%s:", query.ConnID, query.Database, schemaPrefix)
	s.cacheMu.RLock()
	entry, ok := s.tableCache[cacheKey]
	s.cacheMu.RUnlock()
	if ok && time.Since(entry.cachedAt) < s.cacheMaxAge {
		filtered := fuzzyFilterTables(entry.tables, tablePrefix)
		return s.tablesToCandidates(filtered, query.ConnectionType)
	}

	// 缓存未命中则回退到数据库搜索，并后台填充缓存
	go s.prefetchSchemaTables(query.ConnID, query.Database, schemaPrefix)

	q := *query
	q.Schema = schemaPrefix
	q.Prefix = tablePrefix
	return s.completeTableRaw(ctx, &q, tablePrefix)
}

// fuzzyFilterTables 按表名模糊匹配（大小写不敏感子串）。
func fuzzyFilterTables(tables []*schemaservice.TableBO, keyword string) []*schemaservice.TableBO {
	if keyword == "" {
		return tables
	}
	lower := strings.ToLower(keyword)
	result := make([]*schemaservice.TableBO, 0)
	for _, t := range tables {
		if strings.Contains(strings.ToLower(t.Name), lower) {
			result = append(result, t)
		}
	}
	return result
}

// prefetchSchemaTables 后台加载指定 schema 的全部表/视图并缓存。
func (s *ServiceImpl) prefetchSchemaTables(connID int64, database, schema string) {
	if connID <= 0 || database == "" || schema == "" {
		return
	}
	tables, err := s.schemaSvc.ListTables(&schemaservice.ListTablesQuery{
		ConnID:   connID,
		Database: database,
		Schema:   schema,
		Limit:    0,
		Offset:   0,
	})
	if err != nil {
		return
	}
	views, _ := s.schemaSvc.ListViews(
		&schemaservice.ListTablesQuery{
			ConnID:   connID,
			Database: database,
			Schema:   schema,
			Limit:    0,
			Offset:   0,
		})
	for _, v := range views {
		tables = append(tables, &schemaservice.TableBO{Name: v.Name, Type: "VIEW"})
	}
	s.cacheMu.Lock()
	s.tableCache[fmt.Sprintf("%d:%s:%s:", connID, database, schema)] = tableCacheEntry{
		tables:   tables,
		cachedAt: time.Now(),
	}
	s.cacheMu.Unlock()
}

func (s *ServiceImpl) getSchemas(ctx context.Context, query *CompleteQuery) ([]string, error) {
	if query.Database == "" {
		return nil, fmt.Errorf("database is required")
	}
	cacheKey := fmt.Sprintf("%d:%s", query.ConnID, query.Database)
	s.cacheMu.RLock()
	entry, ok := s.schemaCache[cacheKey]
	s.cacheMu.RUnlock()
	if ok && time.Since(entry.cachedAt) < s.cacheMaxAge {
		return entry.schemas, nil
	}

	list, err := s.schemaSvc.ListSchemas(&schemaservice.ListSchemasQuery{
		ConnID:   query.ConnID,
		Database: query.Database,
	})
	if err != nil {
		return nil, err
	}
	schemas := make([]string, 0, len(list))
	for _, sc := range list {
		if sc.Name != "" {
			schemas = append(schemas, sc.Name)
		}
	}

	s.cacheMu.Lock()
	s.schemaCache[cacheKey] = schemaCacheEntry{
		schemas:  schemas,
		cachedAt: time.Now(),
	}
	s.cacheMu.Unlock()
	return schemas, nil
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
// 如果存在多个相同 alias（如自连接或子查询重名），返回所有匹配项，由调用方合并字段。
func (s *ServiceImpl) resolvePrefixToTables(parseRes *parser.ParseResult, query *CompleteQuery) []parser.TableRef {
	prefix := strings.TrimSpace(parseRes.TablePrefix)
	var matched []parser.TableRef
	for _, t := range parseRes.Tables {
		if strings.EqualFold(t.Alias, prefix) {
			matched = append(matched, t)
			continue
		}
		if strings.EqualFold(t.Name, prefix) {
			matched = append(matched, t)
			continue
		}
		// PostgreSQL 可能带 schema 前缀：schema.table
		if t.Schema != "" && strings.EqualFold(t.Schema+"."+t.Name, prefix) {
			matched = append(matched, t)
		}
	}
	if len(matched) > 0 {
		return matched
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
		for k, v := range s.tableCache {
			if now.Sub(v.cachedAt) > s.cacheMaxAge {
				delete(s.tableCache, k)
			}
		}
		for k, v := range s.schemaCache {
			if now.Sub(v.cachedAt) > s.cacheMaxAge {
				delete(s.schemaCache, k)
			}
		}
		s.cacheMu.Unlock()
	}
}

// PrefetchConnectionCache 连接打开后后台预缓存当前数据库的模式/表名。
// 目前主要针对 PostgreSQL，MySQL 保持现状。
func (s *ServiceImpl) PrefetchConnectionCache(connID int64, connectionType, database string) {
	if !isPostgresLike(connectionType) {
		return
	}
	if connID <= 0 || database == "" {
		return
	}
	go func() {
		// 1. 缓存当前数据库的 schema 列表
		schemas, err := s.schemaSvc.ListSchemas(
			&schemaservice.ListSchemasQuery{ConnID: connID, Database: database})
		if err != nil {
			return
		}

		schemaNames := make([]string, 0, len(schemas))
		for _, sc := range schemas {
			if sc.Name != "" {
				schemaNames = append(schemaNames, sc.Name)
			}
		}
		s.cacheMu.Lock()
		s.schemaCache[fmt.Sprintf("%d:%s", connID, database)] = schemaCacheEntry{
			schemas:  schemaNames,
			cachedAt: time.Now(),
		}
		s.cacheMu.Unlock()

		// 2. 缓存当前数据库每个 schema 下的全部表/视图
		for _, sc := range schemas {
			if sc.Name == "" {
				continue
			}
			s.prefetchSchemaTables(connID, database, sc.Name)
		}
	}()
}

// ClearTableCache 清除指定范围的表名缓存。database/schema 为空时按前缀匹配。
func (s *ServiceImpl) ClearTableCache(connID int64, database, schema string) {
	prefix := fmt.Sprintf("%d:", connID)
	if database != "" {
		prefix = fmt.Sprintf("%d:%s:", connID, database)
		if schema != "" {
			prefix = fmt.Sprintf("%d:%s:%s:", connID, database, schema)
		}
	}
	s.cacheMu.Lock()
	for k := range s.tableCache {
		if strings.HasPrefix(k, prefix) {
			delete(s.tableCache, k)
		}
	}
	s.cacheMu.Unlock()
}

// ClearSchemaCache 清除指定范围的 schema 缓存。database 为空时按前缀匹配。
func (s *ServiceImpl) ClearSchemaCache(connID int64, database string) {
	prefix := fmt.Sprintf("%d:", connID)
	if database != "" {
		prefix = fmt.Sprintf("%d:%s:", connID, database)
	}
	s.cacheMu.Lock()
	for k := range s.schemaCache {
		if strings.HasPrefix(k, prefix) {
			delete(s.schemaCache, k)
		}
	}
	s.cacheMu.Unlock()
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
	for k := range s.tableCache {
		if strings.HasPrefix(k, prefix) {
			delete(s.tableCache, k)
		}
	}
	for k := range s.schemaCache {
		if strings.HasPrefix(k, prefix) {
			delete(s.schemaCache, k)
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
