// Package db 提供本地 SQLite 数据源
package db

import (
	"fmt"
	"sync"

	"cyancat/internal/infra/logger"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	database *gorm.DB
	once     sync.Once
	initErr  error
)

// Init 初始化 SQLite 数据库连接
func Init(dbPath string) error {
	once.Do(func() {
		var err error
		database, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "cyancat_",
				SingularTable: false,
			},
			Logger: logger.NewGORMLogger(),
		})
		if err != nil {
			initErr = fmt.Errorf("db: failed to open sqlite: %w", err)
			return
		}

		sqlDB, err := database.DB()
		if err != nil {
			initErr = fmt.Errorf("db: failed to get underlying db: %w", err)
			return
		}

		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)

		logger.L().Info().Str("path", dbPath).Msg("sqlite database initialized")
	})

	return initErr
}

// DB 返回数据库实例
func DB() (*gorm.DB, error) {
	if database == nil {
		return nil, fmt.Errorf("db: not initialized, call Init first")
	}
	return database, nil
}