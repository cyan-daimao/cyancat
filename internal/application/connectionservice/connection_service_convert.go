package connectionservice

import (
	"cyancat/internal/domain/connection"
)

// ToConnectionBO 把 Domain 转换为 BO
func ToConnectionBO(c *connection.Connection) *ConnectionBO {
	if c == nil {
		return nil
	}
	return &ConnectionBO{
		ID:              c.ID,
		Name:            c.Name,
		Type:            c.Type,
		Host:            c.Host,
		Port:            c.Port,
		User:            c.User,
		Password:        c.Password,
		Database:        c.Database,
		SSL:             c.SSL,
		Group:           c.Group,
		Color:           c.Color,
		LastConnectedAt: c.LastConnectedAt,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}

// ToConnectionBOs 批量转换
func ToConnectionBOs(list []*connection.Connection) []*ConnectionBO {
	if len(list) == 0 {
		return make([]*ConnectionBO, 0)
	}
	result := make([]*ConnectionBO, 0, len(list))
	for _, c := range list {
		result = append(result, ToConnectionBO(c))
	}
	return result
}

// ToConnectionFromCreateCmd 把 CreateCmd 转为 Domain
func ToConnectionFromCreateCmd(cmd *CreateConnectionCmd) *connection.Connection {
	if cmd == nil {
		return nil
	}
	return &connection.Connection{
		Name:     cmd.Name,
		Type:     cmd.Type,
		Host:     cmd.Host,
		Port:     cmd.Port,
		User:     cmd.User,
		Password: cmd.Password,
		Database: cmd.Database,
		SSL:      cmd.SSL,
		Group:    cmd.Group,
		Color:    cmd.Color,
	}
}

// ApplyUpdateCmd 把 UpdateCmd 字段写入已存在的 Domain
// 密码为空字符串表示不修改
func ApplyUpdateCmd(c *connection.Connection, cmd *UpdateConnectionCmd) {
	if c == nil || cmd == nil {
		return
	}
	c.Name = cmd.Name
	c.Type = cmd.Type
	c.Host = cmd.Host
	c.Port = cmd.Port
	c.User = cmd.User
	if cmd.Password != "" {
		c.Password = cmd.Password
	}
	c.Database = cmd.Database
	c.SSL = cmd.SSL
	c.Group = cmd.Group
	c.Color = cmd.Color
}

// ToListQuery 把 application 的 ListQuery 转为 domain 的 ListQuery
func ToListQuery(q *ListConnectionQuery) *connection.ListQuery {
	if q == nil {
		return &connection.ListQuery{}
	}
	return &connection.ListQuery{
		Group:   q.Group,
		Type:    q.Type,
		Keyword: q.Keyword,
	}
}

// ToPageQuery 把 application 的 PageQuery 转为 domain 的 PageQuery
func ToPageQuery(q *PageConnectionQuery) *connection.PageQuery {
	if q == nil {
		return &connection.PageQuery{}
	}
	return &connection.PageQuery{
		ListQuery: connection.ListQuery{
			Group:   q.Group,
			Type:    q.Type,
			Keyword: q.Keyword,
		},
		Page:     q.Page,
		PageSize: q.PageSize,
	}
}
