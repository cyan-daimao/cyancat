package mcprepo

import (
	"errors"
	"time"

	"cyancat/internal/infra/db"

	"gorm.io/gorm"
)

// globalConfigID 全局 MCP Server 配置的固定主键
const globalConfigID int64 = 1

// McpServerRepository MCP Server 配置仓储实现（全局单例）
type McpServerRepository struct{}

// NewMcpServerRepository 创建仓储实例
func NewMcpServerRepository() *McpServerRepository {
	return &McpServerRepository{}
}

// GetGlobal 查询全局 MCP Server 配置
func (r *McpServerRepository) GetGlobal() (*McpServerDO, error) {
	database, err := db.DB()
	if err != nil {
		return nil, err
	}

	var do McpServerDO
	if err := database.Where("id = ? AND deleted_at IS NULL", globalConfigID).First(&do).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &do, nil
}

// SaveOrUpdate 保存或更新全局 MCP Server 配置
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
	do.ID = globalConfigID

	// 查找已有记录
	existing, err := r.GetGlobal()
	if err != nil {
		return err
	}
	if existing != nil {
		do.CreatedAt = existing.CreatedAt
		return database.Model(&McpServerDO{}).
			Where("id = ? AND deleted_at IS NULL", globalConfigID).
			Select("enabled", "allow_select", "allow_insert", "allow_update", "allow_delete", "allow_ddl", "port", "token", "updated_at").
			Updates(do).Error
	}

	do.CreatedAt = now
	return database.Create(do).Error
}

// AutoMigrate 自动迁移 MCP Server 表（全局单例模式）
// 如果存在旧的按连接存储的表结构（含 conn_id 列），先删除再重建
func (r *McpServerRepository) AutoMigrate() error {
	database, err := db.DB()
	if err != nil {
		return err
	}

	// 检测旧表结构：如果 mcp_server 表存在且有 conn_id 列，说明是旧版按连接存储的模式
	if database.Migrator().HasTable(&McpServerDO{}) && database.Migrator().HasColumn(&McpServerDO{}, "conn_id") {
		// 删除旧表，重建为全局单例模式
		if err := database.Migrator().DropTable(&McpServerDO{}); err != nil {
			return err
		}
	}

	return database.AutoMigrate(&McpServerDO{})
}
