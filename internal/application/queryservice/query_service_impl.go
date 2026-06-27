package queryservice

import (
	"context"
	"errors"
	"strings"
	"time"

	"cyancat/internal/infra/api"
	"cyancat/internal/infra/driver"
	"cyancat/internal/infra/eventbus"
	"cyancat/internal/infra/logger"
	"cyancat/internal/infra/session"
)

// QueryServiceImpl 查询服务实现
type QueryServiceImpl struct {
	sessionMgr  session.Manager
	bus         eventbus.Bus
	historyRepo HistoryRepository
}

// NewQueryServiceImpl 创建查询服务
func NewQueryServiceImpl(sessionMgr session.Manager, bus eventbus.Bus, historyRepo HistoryRepository) *QueryServiceImpl {
	return &QueryServiceImpl{
		sessionMgr:  sessionMgr,
		bus:         bus,
		historyRepo: historyRepo,
	}
}

// Execute 执行 SQL
func (s *QueryServiceImpl) Execute(cmd *ExecuteCmd) (*QueryResultBO, error) {
	if cmd == nil {
		return nil, errors.New("queryservice: cmd cannot be nil")
	}
	if cmd.ConnID <= 0 {
		return nil, errors.New("queryservice: connID must be positive")
	}

	sqlText := strings.TrimSpace(cmd.SQL)
	if sqlText == "" {
		return nil, errors.New("queryservice: sql cannot be empty")
	}

	conn, err := s.sessionMgr.Get(cmd.ConnID)
	if err != nil {
		return nil, err
	}

	maxRows := cmd.MaxRows
	if maxRows <= 0 {
		maxRows = 1000
	}

	// 多语句拆分：用分号拆分，仅取最后一条非空语句作为主结果
	stmts := splitStatements(sqlText)
	if len(stmts) == 0 {
		return nil, errors.New("queryservice: no executable statement")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var (
		result *QueryResultBO
		start  = time.Now()
	)

	execConn, cleanup, err := s.prepareExecutionConn(ctx, cmd, conn)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		s.recordError(cmd.ConnID, sqlText, err)
		if s.historyRepo != nil {
			_ = s.historyRepo.Save(&QueryHistoryBO{
				ConnID:       cmd.ConnID,
				SQL:          sqlText,
				Status:       "error",
				ErrorMessage: err.Error(),
				DurationMs:   time.Since(start).Milliseconds(),
				ExecutedAt:   time.Now(),
			})
		}
		return nil, err
	}

	// 如果所有语句都是非查询（DDL/DML/SET/COMMENT 等），将整段 SQL 作为 batch 发送。
	// 这样 pgx 会自动走简单查询协议，正确处理 SET search_path、COMMENT ON 等
	// 扩展协议下无法执行的工具类语句。
	// 注意：仅对 PostgreSQL 启用 batch 模式；MySQL 驱动默认不支持 multiStatements。
	driverType, _ := s.sessionMgr.DriverType(cmd.ConnID)
	if driverType == driver.PostgreSQL && allNonQuery(stmts) {
		res, err := execConn.Execute(ctx, sqlText)
		if err != nil {
			s.recordError(cmd.ConnID, sqlText, err)
			if s.historyRepo != nil {
				_ = s.historyRepo.Save(&QueryHistoryBO{
					ConnID:       cmd.ConnID,
					SQL:          sqlText,
					Status:       "error",
					ErrorMessage: err.Error(),
					DurationMs:   time.Since(start).Milliseconds(),
					ExecutedAt:   time.Now(),
				})
			}
			return nil, err
		}
		result = &QueryResultBO{
			ConnID:       cmd.ConnID,
			SQL:          sqlText,
			Columns:      ToColumnBOs(res.Columns),
			Rows:         res.Rows,
			RowsAffected: res.RowsAffected,
			LastInsertID: res.LastInsertID,
			Duration:     time.Since(start),
		}
	} else {
		// 包含查询语句：逐条执行，仅最后一条返回结果，且自动追加 LIMIT 兜底
		for idx, stmt := range stmts {
			isLast := idx == len(stmts)-1
			stmtSQL := stripTrailingSemicolon(stmt)
			if stmtSQL == "" {
				continue
			}

			finalSQL := stmtSQL
			truncated := false
			if isLast && shouldAddLimit(stmtSQL) {
				finalSQL = stmtSQL + " LIMIT " + itoa(maxRows)
				truncated = true
			}

			res, err := execConn.Execute(ctx, finalSQL)
			if err != nil {
				s.recordError(cmd.ConnID, stmtSQL, err)
				if s.historyRepo != nil {
					_ = s.historyRepo.Save(&QueryHistoryBO{
						ConnID:       cmd.ConnID,
						SQL:          stmtSQL,
						Status:       "error",
						ErrorMessage: err.Error(),
						DurationMs:   time.Since(start).Milliseconds(),
						ExecutedAt:   time.Now(),
					})
				}
				return nil, err
			}

			if !isLast {
				continue
			}

			result = &QueryResultBO{
				ConnID:       cmd.ConnID,
				SQL:          finalSQL,
				Columns:      ToColumnBOs(res.Columns),
				Rows:         res.Rows,
				RowsAffected: res.RowsAffected,
				LastInsertID: res.LastInsertID,
				Duration:     time.Since(start),
				Truncated:    truncated && int64(len(res.Rows)) >= int64(maxRows),
			}
		}
	}

	if result == nil {
		return nil, errors.New("queryservice: no result produced")
	}

	logger.L().Info().
		Int64("connID", cmd.ConnID).
		Int("rows", len(result.Rows)).
		Dur("duration", result.Duration).
		Msg("query executed")

	// 记录成功历史
	if s.historyRepo != nil {
		historyBO := &QueryHistoryBO{
			ConnID:     cmd.ConnID,
			SQL:        sqlText,
			Status:     "success",
			RowCount:   int64(len(result.Rows)),
			DurationMs: result.Duration.Milliseconds(),
			ExecutedAt: time.Now(),
		}
		_ = s.historyRepo.Save(historyBO)
	}

	return result, nil
}

// Cancel 取消查询（V1.5 实现，当前空实现）
func (s *QueryServiceImpl) Cancel(connID int64) error {
	// TODO V1.5：维护 connID -> context.CancelFunc 映射
	return nil
}

// History 查询历史
func (s *QueryServiceImpl) History(query *HistoryQuery) (*api.Page[*QueryHistoryBO], error) {
	if query == nil {
		query = &HistoryQuery{}
	}
	if s.historyRepo == nil {
		return api.NewPage[*QueryHistoryBO](make([]*QueryHistoryBO, 0), 0, query.Page, query.PageSize), nil
	}
	list, total, err := s.historyRepo.Page(query)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*QueryHistoryBO, 0)
	}
	return api.NewPage(list, total, query.Page, query.PageSize), nil
}

func (s *QueryServiceImpl) prepareExecutionConn(ctx context.Context, cmd *ExecuteCmd, conn driver.Conn) (driver.Conn, func(), error) {
	database := strings.TrimSpace(cmd.Database)
	schema := strings.TrimSpace(cmd.Schema)
	if database == "" && schema == "" {
		return conn, func() {}, nil
	}

	driverType, err := s.sessionMgr.DriverType(cmd.ConnID)
	if err != nil {
		return nil, nil, err
	}

	switch driverType {
	case driver.MySQL:
		if database == "" {
			database = schema
		}
		if database == "" {
			return conn, func() {}, nil
		}
		_, err := conn.Execute(ctx, "USE "+quoteMySQLIdent(database))
		if err != nil {
			return nil, nil, err
		}
		return conn, func() {}, nil
	case driver.PostgreSQL:
		execConn, cleanup, err := conn.WithDatabase(ctx, database)
		if err != nil {
			return nil, nil, err
		}
		if schema == "" {
			return execConn, cleanup, nil
		}
		if _, err := execConn.Execute(ctx, "SET search_path TO "+quotePostgresIdent(schema)); err != nil {
			cleanup()
			return nil, nil, err
		}
		return execConn, cleanup, nil
	default:
		return conn, func() {}, nil
	}
}

// recordError 记录执行错误到 EventBus
func (s *QueryServiceImpl) recordError(connID int64, sqlText string, err error) {
	if s.bus == nil {
		return
	}
	s.bus.Emit(eventbus.EventQueryError, map[string]any{
		"connID": connID,
		"sql":    sqlText,
		"error":  err.Error(),
	})
}

// --- 辅助 ---

// splitStatements 按 ; 拆分 SQL 语句，正确处理字符串、标识符引用和注释中的分号。
func splitStatements(sqlText string) []string {
	var (
		result []string
		start  = 0
		i      = 0
		n      = len(sqlText)
	)

	for i < n {
		c := sqlText[i]

		switch {
		// 单引号字符串：跳过 '' 转义
		case c == '\'':
			i++ // skip opening quote
			for i < n {
				if sqlText[i] == '\'' {
					// 检查是否是转义 ''
					if i+1 < n && sqlText[i+1] == '\'' {
						i += 2 // skip escaped ''
						continue
					}
					i++ // skip closing quote
					break
				}
				i++
			}

		// 双引号标识符：跳过 "" 转义
		case c == '"':
			i++ // skip opening quote
			for i < n {
				if sqlText[i] == '"' {
					if i+1 < n && sqlText[i+1] == '"' {
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}

		// PostgreSQL 美元引用: $tag$...$tag$
		case c == '$':
			// 提取 tag（可能为空）
			tagStart := i + 1
			for tagStart < n && sqlText[tagStart] != '$' {
				tagStart++
			}
			if tagStart < n { // 找到配对的 $，即 $tag$
				tagEnd := tagStart + 1 // 跳过闭合 $
				tag := sqlText[i+1 : tagStart]
				// 查找结束标记 $tag$
				closeTag := "$" + tag + "$"
				closeIdx := strings.Index(sqlText[tagEnd:], closeTag)
				if closeIdx >= 0 {
					i = tagEnd + closeIdx + len(closeTag)
					continue
				}
			}
			i++

		// 行注释 --
		case c == '-' && i+1 < n && sqlText[i+1] == '-':
			// 跳到行尾
			i += 2
			for i < n && sqlText[i] != '\n' && sqlText[i] != '\r' {
				i++
			}

		// 块注释 /*
		case c == '/' && i+1 < n && sqlText[i+1] == '*':
			i += 2
			for i+1 < n {
				if sqlText[i] == '*' && sqlText[i+1] == '/' {
					i += 2
					break
				}
				i++
			}

		// 分号 = 语句边界
		case c == ';':
			part := strings.TrimSpace(sqlText[start:i])
			if part != "" {
				result = append(result, part)
			}
			i++
			start = i

		default:
			i++
		}
	}

	// 最后一段（末尾没有分号的情况）
	part := strings.TrimSpace(sqlText[start:])
	if part != "" {
		result = append(result, part)
	}

	return result
}

// stripTrailingSemicolon 去除末尾分号及多余空白
func stripTrailingSemicolon(sql string) string {
	s := strings.TrimSpace(sql)
	if strings.HasSuffix(s, ";") {
		s = strings.TrimSuffix(s, ";")
		s = strings.TrimSpace(s)
	}
	return s
}

// allNonQuery 判断所有语句是否都是非查询（DDL/DML/SET/COMMENT 等不返回结果集的语句）。
// 当所有语句都是非查询时，整段 SQL 可作为 batch 发送，让 pgx 走简单查询协议，
// 从而正确执行 SET search_path、COMMENT ON 等扩展协议不支持的工具类语句。
func allNonQuery(stmts []string) bool {
	for _, stmt := range stmts {
		trimmed := skipLeadingNoise(stmt)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		for _, prefix := range []string{"select", "with", "show", "desc", "explain"} {
			if strings.HasPrefix(lower, prefix) {
				return false
			}
		}
	}
	return true
}

// skipLeadingNoise 跳过 SQL 语句前面的空白和注释，返回第一个有效关键字开始的部分。
func skipLeadingNoise(s string) string {
	s = strings.TrimSpace(s)
	for len(s) > 0 {
		switch {
		case strings.HasPrefix(s, "--"):
			idx := strings.IndexByte(s, '\n')
			if idx < 0 {
				return ""
			}
			s = strings.TrimSpace(s[idx+1:])
		case strings.HasPrefix(s, "/*"):
			idx := strings.Index(s, "*/")
			if idx < 0 {
				return ""
			}
			s = strings.TrimSpace(s[idx+2:])
		default:
			return s
		}
	}
	return s
}

// shouldAddLimit 判断是否需要自动追加 LIMIT（仅对没有 LIMIT 的 SELECT）
func shouldAddLimit(sqlText string) bool {
	lower := strings.ToLower(strings.TrimSpace(sqlText))
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return false
	}
	if strings.Contains(lower, " limit ") || strings.HasSuffix(lower, " limit") {
		return false
	}
	return true
}

// itoa 整数转字符串（避免 strconv import）
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func quoteMySQLIdent(ident string) string {
	var b strings.Builder
	b.Grow(len(ident) + 2)
	b.WriteByte('`')
	for i := 0; i < len(ident); i++ {
		if ident[i] == '`' {
			b.WriteString("``")
		} else {
			b.WriteByte(ident[i])
		}
	}
	b.WriteByte('`')
	return b.String()
}

func quotePostgresIdent(ident string) string {
	var b strings.Builder
	b.Grow(len(ident) + 2)
	b.WriteByte('"')
	for i := 0; i < len(ident); i++ {
		if ident[i] == '"' {
			b.WriteString(`""`)
		} else {
			b.WriteByte(ident[i])
		}
	}
	b.WriteByte('"')
	return b.String()
}
