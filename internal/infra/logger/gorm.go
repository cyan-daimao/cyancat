package logger

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormZerologAdapter 把 zerolog 适配到 gorm Logger 接口
type gormZerologAdapter struct {
	slowThreshold time.Duration
}

// NewGORMLogger 创建供 GORM 使用的 Logger
func NewGORMLogger() gormlogger.Interface {
	return &gormZerologAdapter{
		slowThreshold: 200 * time.Millisecond,
	}
}

// LogMode 兼容 GORM 日志接口
func (l *gormZerologAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

// Info 输出信息日志
func (l *gormZerologAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	L().Info().Fields(data).Msg(msg)
}

// Warn 输出警告日志
func (l *gormZerologAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	L().Warn().Fields(data).Msg(msg)
}

// Error 输出错误日志
func (l *gormZerologAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	L().Error().Fields(data).Msg(msg)
}

// Trace 输出 SQL 跟踪日志
func (l *gormZerologAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	event := L().Debug()
	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		event = L().Error().Err(err)
	case elapsed > l.slowThreshold:
		event = L().Warn().Str("slow", "true")
	}

	event.
		Dur("elapsed", elapsed).
		Int64("rows", rows).
		Str("sql", sql).
		Msg("gorm trace")
}