package postgres

import (
	"fmt"

	"cyancat/internal/infra/driver"

	"github.com/jackc/pgx/v5/pgtype"
)

// oidTypeName 把 PG OID 转可读类型名（简化版，覆盖常用类型）
func oidTypeName(oid uint32) string {
	switch oid {
	case pgtype.BoolOID:
		return "bool"
	case pgtype.Int2OID:
		return "int2"
	case pgtype.Int4OID:
		return "int4"
	case pgtype.Int8OID:
		return "int8"
	case pgtype.Float4OID:
		return "float4"
	case pgtype.Float8OID:
		return "float8"
	case pgtype.NumericOID:
		return "numeric"
	case pgtype.TextOID:
		return "text"
	case pgtype.VarcharOID:
		return "varchar"
	case pgtype.BPCharOID:
		return "char"
	case pgtype.ByteaOID:
		return "bytea"
	case pgtype.DateOID:
		return "date"
	case pgtype.TimestampOID:
		return "timestamp"
	case pgtype.TimestamptzOID:
		return "timestamptz"
	case pgtype.JSONOID:
		return "json"
	case pgtype.JSONBOID:
		return "jsonb"
	case pgtype.UUIDOID:
		return "uuid"
	default:
		return fmt.Sprintf("oid_%d", oid)
	}
}

// scanAllPG 收集所有行
func scanAllPG(rows pgxRowsLike, colCount int) ([][]any, error) {
	var result [][]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			values[i] = driver.NormalizeValue(v)
		}
		result = append(result, values)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// pgxRowsLike 抽出方法集，便于复用（避免 import cycle）
type pgxRowsLike interface {
	Next() bool
	Values() ([]any, error)
	Err() error
}
