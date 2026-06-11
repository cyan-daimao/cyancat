package connectionrepo

import (
	"cyancat/internal/domain/connection"
	"cyancat/internal/infra/driver"
)

// ToConnection 把 DO 转换为 Domain（仅模型字段，密码由调用方按 keychain 解密回填）
func ToConnection(do *ConnectionDO) *connection.Connection {
	if do == nil {
		return nil
	}
	return &connection.Connection{
		ID:              do.ID,
		Name:            do.Name,
		Type:            driver.DriverType(do.Type),
		Host:            do.Host,
		Port:            do.Port,
		User:            do.User,
		// Password 留空，由 application 层从 keychain 读取后注入
		Password:        "",
		Database:        do.Database,
		SSL:             do.SSL,
		Group:           do.Group,
		Color:           do.Color,
		LastConnectedAt: do.LastConnectedAt,
		CreatedAt:       do.CreatedAt,
		UpdatedAt:       do.UpdatedAt,
		DeletedAt:       do.DeletedAt,
	}
}

// ToConnections 批量转换
func ToConnections(dos []ConnectionDO) []*connection.Connection {
	if len(dos) == 0 {
		return nil
	}
	result := make([]*connection.Connection, 0, len(dos))
	for i := range dos {
		result = append(result, ToConnection(&dos[i]))
	}
	return result
}

// ToConnectionDO 把 Domain 转换为 DO
// passwordEncrypted 由 application 层先用 keychain 加密后传入
func ToConnectionDO(c *connection.Connection, passwordEncrypted string) *ConnectionDO {
	if c == nil {
		return nil
	}
	return &ConnectionDO{
		ID:                c.ID,
		Name:              c.Name,
		Type:              c.Type.String(),
		Host:              c.Host,
		Port:              c.Port,
		User:              c.User,
		PasswordEncrypted: passwordEncrypted,
		Database:          c.Database,
		SSL:               c.SSL,
		Group:             c.Group,
		Color:             c.Color,
		LastConnectedAt:   c.LastConnectedAt,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
		DeletedAt:         c.DeletedAt,
	}
}
