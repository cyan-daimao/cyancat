package mcprepo

import (
	"errors"
	"time"

	"cyancat/internal/infra/db"

	"gorm.io/gorm"
)

// McpServerRepository MCP Server 配置仓储实现
type McpServerRepository struct{}

// NewMcpServerRepository 创建仓储实例
func NewMcpServerRepository() *McpServerRepository {
	return &McpServerRepository{}
}

// GetByConnID 按连接 ID 查询 MCP Server 配置
func (r *McpServerRepository) GetByConnID(connID int64) (*McpServerDO, error) {
	database, err := db.DB()
	if err != nil {
		return nil, err
	}

	var do McpServerDO
	if err := database.Where("conn_id = ? AND deleted_at IS NULL", connID).First(&do).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &do, nil
}

// SaveOrUpdate 保存或更新 MCP Server 配置
func (r *McpServerRepository) SaveOrUpdate(do *McpServerDO) error {
	if do == nil {
		return errors.New("mcprepo: do cannot be nil")
	}

	database, err := db.DB()
	if err != nil {
		return err
	}

	now := time.Now()
	do.UpdatedAt = now

	// 按 conn_id 查找已有记录，避免唯一索引冲突
	existing, err := r.GetByConnID(do.ConnID)
	if err != nil {
		return err
	}
	if existing != nil {
		do.ID = existing.ID
		do.CreatedAt = existing.CreatedAt
		return database.Model(&McpServerDO{}).
			Where("id = ? AND deleted_at IS NULL", do.ID).
			Select("conn_id", "enabled", "allow_select", "allow_insert", "allow_update", "allow_delete", "allow_ddl", "port", "token", "updated_at").
			Updates(do).Error
	}

	do.CreatedAt = now
	return database.Create(do).Error
}

// AutoMigrate 自动迁移 MCP Server 表
func (r *McpServerRepository) AutoMigrate() error {
	database, err := db.DB()
	if err != nil {
		return err
	}
	return database.AutoMigrate(&McpServerDO{})
}
