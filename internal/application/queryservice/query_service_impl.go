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

	// 多语句拆分：仅取最后一条非空语句作为主结果（其它语句也执行，但只返回最后一条的结果）
	// V1.0 简化版：用分号粗暴拆分，未来用真正的 SQL parser
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

	// 前 N-1 条用 Execute（忽略结果集），最后一条返回给前端
	for idx, stmt := range stmts {
		isLast := idx == len(stmts)-1
		stmtSQL := stripTrailingSemicolon(stmt)
		if stmtSQL == "" {
			continue
		}

		// 最后一条带 LIMIT 兜底
		finalSQL := stmtSQL
		truncated := false
		if isLast && shouldAddLimit(stmtSQL) {
			finalSQL = stmtSQL + " LIMIT " + itoa(maxRows)
			truncated = true
		}

		res, err := execConn.Execute(ctx, finalSQL)
		if err != nil {
			s.recordError(cmd.ConnID, stmtSQL, err)
			// 记录失败历史
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

// splitStatements 按 ; 粗暴拆分 SQL（V1.0 简化版，不处理字符串/注释中的分号）
func splitStatements(sqlText string) []string {
	parts := strings.Split(sqlText, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			result = append(result, p)
		}
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
