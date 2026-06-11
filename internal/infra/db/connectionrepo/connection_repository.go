package connectionrepo

import (
	"errors"
	"strings"
	"time"

	"cyancat/internal/domain/connection"
	"cyancat/internal/infra/crypto"
	"cyancat/internal/infra/db"

	"gorm.io/gorm"
)

// ConnectionRepository 数据库连接仓储实现
type ConnectionRepository struct {
	// masterKey AES 主密钥（由 keychain 在启动时注入）
	masterKey []byte
}

// NewConnectionRepository 创建仓储实例
// masterKey 必须为 32 字节
func NewConnectionRepository(masterKey []byte) *ConnectionRepository {
	return &ConnectionRepository{
		masterKey: masterKey,
	}
}

// List 列出连接
func (r *ConnectionRepository) List(query *connection.ListQuery) ([]*connection.Connection, error) {
	database, err := db.DB()
	if err != nil {
		return nil, err
	}

	tx := database.Model(&ConnectionDO{}).Where("deleted_at IS NULL")
	tx = applyFilter(tx, query)

	var dos []ConnectionDO
	if err := tx.Order("id DESC").Find(&dos).Error; err != nil {
		return nil, err
	}

	return r.decryptAndConvert(dos)
}

// Page 分页查询
func (r *ConnectionRepository) Page(query *connection.PageQuery) ([]*connection.Connection, int64, error) {
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
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	tx := database.Model(&ConnectionDO{}).Where("deleted_at IS NULL")
	tx = applyFilter(tx, &query.ListQuery)

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var dos []ConnectionDO
	if err := tx.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&dos).Error; err != nil {
		return nil, 0, err
	}

	list, err := r.decryptAndConvert(dos)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetByID 按 ID 查询
func (r *ConnectionRepository) GetByID(id int64) (*connection.Connection, error) {
	database, err := db.DB()
	if err != nil {
		return nil, err
	}

	var do ConnectionDO
	if err := database.Where("id = ? AND deleted_at IS NULL", id).First(&do).Error; err != nil {
		return nil, err
	}

	c := ToConnection(&do)
	if err := r.decryptInto(&do, c); err != nil {
		return nil, err
	}
	return c, nil
}

// GetByName 按名称查询（用于唯一性校验）
func (r *ConnectionRepository) GetByName(name string) (*connection.Connection, error) {
	database, err := db.DB()
	if err != nil {
		return nil, err
	}

	var do ConnectionDO
	if err := database.Where("name = ? AND deleted_at IS NULL", strings.TrimSpace(name)).
		First(&do).Error; err != nil {
		return nil, err
	}

	c := ToConnection(&do)
	if err := r.decryptInto(&do, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Save 新建连接
func (r *ConnectionRepository) Save(conn *connection.Connection) error {
	if conn == nil {
		return errors.New("connection: cannot be nil")
	}
	database, err := db.DB()
	if err != nil {
		return err
	}

	encrypted, err := r.encrypt(conn.Password)
	if err != nil {
		return err
	}

	do := ToConnectionDO(conn, encrypted)
	if do == nil {
		return errors.New("connection: cannot be nil")
	}

	if err := database.Create(do).Error; err != nil {
		return err
	}

	saved := ToConnection(do)
	// 保留原密码明文（Save 后业务可能直接复用）
	saved.Password = conn.Password
	*conn = *saved
	return nil
}

// Update 更新连接
func (r *ConnectionRepository) Update(conn *connection.Connection) error {
	if conn == nil {
		return errors.New("connection: cannot be nil")
	}
	if conn.ID <= 0 {
		return errors.New("connection: id must be positive")
	}
	database, err := db.DB()
	if err != nil {
		return err
	}

	encrypted, err := r.encrypt(conn.Password)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"name":               conn.Name,
		"type":               conn.Type.String(),
		"host":               conn.Host,
		"port":               conn.Port,
		"user":               conn.User,
		"password_encrypted": encrypted,
		"database":           conn.Database,
		"ssl":                conn.SSL,
		"group":              conn.Group,
		"color":              conn.Color,
		"last_connected_at":  conn.LastConnectedAt,
		"updated_at":         time.Now(),
		"updated_by":         "",
	}

	result := database.Model(&ConnectionDO{}).
		Where("id = ? AND deleted_at IS NULL", conn.ID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 软删除连接
func (r *ConnectionRepository) Delete(id int64) error {
	database, err := db.DB()
	if err != nil {
		return err
	}

	now := time.Now()
	result := database.Model(&ConnectionDO{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", &now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// --- 私有辅助方法 ---

func (r *ConnectionRepository) encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	if len(r.masterKey) != 32 {
		return "", errors.New("connectionrepo: master key not initialized")
	}
	return crypto.Encrypt([]byte(plain), r.masterKey)
}

func (r *ConnectionRepository) decrypt(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}
	if len(r.masterKey) != 32 {
		return "", errors.New("connectionrepo: master key not initialized")
	}
	raw, err := crypto.Decrypt(encrypted, r.masterKey)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (r *ConnectionRepository) decryptInto(do *ConnectionDO, c *connection.Connection) error {
	if do.PasswordEncrypted == "" {
		return nil
	}
	plain, err := r.decrypt(do.PasswordEncrypted)
	if err != nil {
		return err
	}
	c.Password = plain
	return nil
}

func (r *ConnectionRepository) decryptAndConvert(dos []ConnectionDO) ([]*connection.Connection, error) {
	result := make([]*connection.Connection, 0, len(dos))
	for i := range dos {
		c := ToConnection(&dos[i])
		if err := r.decryptInto(&dos[i], c); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

func applyFilter(tx *gorm.DB, query *connection.ListQuery) *gorm.DB {
	if query == nil {
		return tx
	}
	if query.Group != "" {
		tx = tx.Where("`group` = ?", query.Group)
	}
	if query.Type != "" {
		tx = tx.Where("type = ?", query.Type)
	}
	if query.Keyword != "" {
		like := "%" + strings.TrimSpace(query.Keyword) + "%"
		tx = tx.Where("name LIKE ? OR host LIKE ?", like, like)
	}
	return tx
}
