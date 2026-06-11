package connectionrepo

import (
	"cyancat/internal/infra/db"
)

// AutoMigrate 自动迁移 ConnectionDO 对应的表
func AutoMigrate() error {
	database, err := db.DB()
	if err != nil {
		return err
	}
	return database.AutoMigrate(&ConnectionDO{})
}
