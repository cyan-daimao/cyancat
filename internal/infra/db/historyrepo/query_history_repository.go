package historyrepo

import (
	"errors"
	"strings"
	"time"

	"cyancat/internal/application/queryservice"
	"cyancat/internal/infra/db"
)

// QueryHistoryRepository 查询历史仓储实现
type QueryHistoryRepository struct{}

// NewQueryHistoryRepository 创建实例
func NewQueryHistoryRepository() *QueryHistoryRepository {
	return &QueryHistoryRepository{}
}

// Save 保存一条历史
func (r *QueryHistoryRepository) Save(bo *queryservice.QueryHistoryBO) error {
	if bo == nil {
		return errors.New("historyrepo: bo cannot be nil")
	}
	database, err := db.DB()
	if err != nil {
		return err
	}

	do := ToQueryHistoryDO(bo)
	if do.CreatedAt.IsZero() {
		do.CreatedAt = time.Now()
	}
	do.UpdatedAt = do.CreatedAt

	if err := database.Create(do).Error; err != nil {
		return err
	}
	bo.ID = do.ID
	return nil
}

// Page 分页查询历史
func (r *QueryHistoryRepository) Page(query *queryservice.HistoryQuery) ([]*queryservice.QueryHistoryBO, int64, error) {
	database, err := db.DB()
	if err != nil {
		return nil, 0, err
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	tx := database.Model(&QueryHistoryDO{}).Where("deleted_at IS NULL")
	if query.ConnID > 0 {
		tx = tx.Where("conn_id = ?", query.ConnID)
	}
	if query.Status != "" {
		tx = tx.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		like := "%" + strings.TrimSpace(query.Keyword) + "%"
		tx = tx.Where("sql LIKE ?", like)
	}
	if query.StartTime != nil {
		tx = tx.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		tx = tx.Where("created_at <= ?", *query.EndTime)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []QueryHistoryDO
	if err := tx.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return ToQueryHistoryBOs(list), total, nil
}

// DeleteBefore 批量软删除指定时间之前的历史
func (r *QueryHistoryRepository) DeleteBefore(before time.Time) error {
	database, err := db.DB()
	if err != nil {
		return err
	}
	now := time.Now()
	return database.Model(&QueryHistoryDO{}).
		Where("deleted_at IS NULL AND created_at < ?", before).
		Update("deleted_at", &now).Error
}

// AutoMigrate 自动迁移历史表
func AutoMigrate() error {
	database, err := db.DB()
	if err != nil {
		return err
	}
	return database.AutoMigrate(&QueryHistoryDO{})
}
