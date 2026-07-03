package driver

import (
	"math/big"
	"strconv"
)

// maxSafeInteger 是 JavaScript 中 Number 类型能精确表示的最大整数（2^53 - 1）。
// 超过该范围的整数在 JSON -> JS 反序列化时会丢失精度。
const maxSafeInteger = 9007199254740991

// NormalizeValue 把可能超出 JS 安全整数范围的数值类型转为 string，
// 避免前端/Wails JSON 反序列化后精度失真。
// 已在安全范围内的整数保持原类型，避免影响普通 int 的展示与比较。
func NormalizeValue(v any) any {
	switch n := v.(type) {
	case int:
		if n > maxSafeInteger || n < -maxSafeInteger {
			return strconv.FormatInt(int64(n), 10)
		}
	case int8:
		return int16(n)
	case int16:
		return n
	case int32:
		return n
	case int64:
		if n > maxSafeInteger || n < -maxSafeInteger {
			return strconv.FormatInt(n, 10)
		}
	case uint:
		if uint64(n) > maxSafeInteger {
			return strconv.FormatUint(uint64(n), 10)
		}
	case uint8:
		return uint16(n)
	case uint16:
		return n
	case uint32:
		if uint64(n) > maxSafeInteger {
			return strconv.FormatUint(uint64(n), 10)
		}
	case uint64:
		if n > maxSafeInteger {
			return strconv.FormatUint(n, 10)
		}
	case big.Int:
		return n.String()
	case *big.Int:
		if n == nil {
			return nil
		}
		return n.String()
	}
	return v
}

// NormalizeRows 对二维结果集逐格执行 NormalizeValue。
func NormalizeRows(rows [][]any) [][]any {
	for i, row := range rows {
		for j, v := range row {
			row[j] = NormalizeValue(v)
		}
		rows[i] = row
	}
	return rows
}
