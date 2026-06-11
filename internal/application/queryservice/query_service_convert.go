package queryservice

import "cyancat/internal/infra/driver"

// ToColumnBOs 把 driver.Column 转 BO
func ToColumnBOs(cols []driver.Column) []ColumnBO {
	if len(cols) == 0 {
		return make([]ColumnBO, 0)
	}
	result := make([]ColumnBO, 0, len(cols))
	for _, c := range cols {
		result = append(result, ColumnBO{
			Name:         c.Name,
			DatabaseType: c.DatabaseType,
			Nullable:     c.Nullable,
		})
	}
	return result
}
