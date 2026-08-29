// 数据库类型工具：按连接类型提供字段类型候选与类型归一化

// MySQL 常用字段类型
export const MYSQL_TYPES = [
  'INT', 'BIGINT', 'SMALLINT', 'TINYINT',
  'VARCHAR', 'CHAR', 'TEXT', 'LONGTEXT', 'MEDIUMTEXT',
  'DATETIME', 'TIMESTAMP', 'DATE', 'TIME',
  'DECIMAL', 'FLOAT', 'DOUBLE',
  'JSON', 'BLOB', 'BOOLEAN', 'ENUM',
];

// PostgreSQL 常用字段类型
export const PG_TYPES = [
  'INTEGER', 'BIGINT', 'SMALLINT',
  'SERIAL', 'BIGSERIAL', 'SMALLSERIAL',
  'VARCHAR', 'CHAR', 'TEXT',
  'NUMERIC', 'DECIMAL', 'REAL', 'DOUBLE PRECISION',
  'BOOLEAN',
  'DATE', 'TIME', 'TIMESTAMP', 'TIMESTAMPTZ',
  'JSON', 'JSONB', 'UUID', 'BYTEA',
];

// SQLite 常用字段类型
export const SQLITE_TYPES = [
  'INTEGER', 'REAL', 'TEXT', 'BLOB', 'NUMERIC',
];

/** 按连接类型返回字段类型候选列表 */
export function getCommonTypes(connectionType?: string): string[] {
  switch (connectionType) {
    case 'postgres':
      return PG_TYPES;
    case 'sqlite':
      return SQLITE_TYPES;
    case 'mysql':
    case 'starrocks':
    default:
      return MYSQL_TYPES;
  }
}

/** 判断该连接类型是否支持 UNSIGNED 修饰 */
export function supportsUnsigned(connectionType?: string): boolean {
  return connectionType === 'mysql' || connectionType === 'starrocks';
}

/** 判断该连接类型是否支持表级 engine/charset/collation 选项 */
export function supportsTableOptions(connectionType?: string): boolean {
  return connectionType === 'mysql' || connectionType === 'starrocks';
}

/**
 * 归一化数据库返回的类型字符串到 UI 候选值
 * MySQL: 'varchar(255)' → 'VARCHAR', 'int(11) unsigned' → 'INT'
 */
export function normalizeDataType(raw: string): string {
  if (!raw) return 'INT';
  let s = raw.replace(/\(.*?\)/g, '');
  s = s.replace(/\b(unsigned|zerofill|signed)\b/gi, '');
  return s.trim().toUpperCase();
}

/**
 * 归一化 PostgreSQL format_type 输出到 PG_TYPES 候选值
 * 例：'character varying(255)' → 'VARCHAR'
 *     'integer' → 'INTEGER'
 *     'timestamp without time zone' → 'TIMESTAMP'
 *     'numeric(10,2)' → 'NUMERIC'
 */
export function normalizeDataTypeForPG(raw: string): string {
  if (!raw) return 'INTEGER';
  // 去掉括号及其内容
  let s = raw.replace(/\(.*?\)/g, '').trim().toLowerCase();

  // 精确匹配常见 format_type 输出
  switch (s) {
    case 'character varying':
    case 'varchar':
      return 'VARCHAR';
    case 'character':
    case 'char':
    case 'bpchar':
      return 'CHAR';
    case 'text':
      return 'TEXT';
    case 'integer':
    case 'int':
    case 'int4':
      return 'INTEGER';
    case 'bigint':
    case 'int8':
      return 'BIGINT';
    case 'smallint':
    case 'int2':
      return 'SMALLINT';
    case 'numeric':
    case 'decimal':
      return 'NUMERIC';
    case 'real':
    case 'float4':
      return 'REAL';
    case 'double precision':
    case 'float8':
      return 'DOUBLE PRECISION';
    case 'boolean':
    case 'bool':
      return 'BOOLEAN';
    case 'date':
      return 'DATE';
    case 'time':
    case 'time without time zone':
      return 'TIME';
    case 'timestamp':
    case 'timestamp without time zone':
      return 'TIMESTAMP';
    case 'timestamptz':
    case 'timestamp with time zone':
      return 'TIMESTAMPTZ';
    case 'json':
      return 'JSON';
    case 'jsonb':
      return 'JSONB';
    case 'uuid':
      return 'UUID';
    case 'bytea':
      return 'BYTEA';
    case 'serial':
    case 'serial4':
      return 'SERIAL';
    case 'bigserial':
    case 'serial8':
      return 'BIGSERIAL';
    case 'smallserial':
    case 'serial2':
      return 'SMALLSERIAL';
  }

  // 未识别的类型：尝试大写后看是否在候选列表中
  const upper = s.toUpperCase();
  if (PG_TYPES.includes(upper)) {
    return upper;
  }
  // fallback 到 TEXT，并保留原值（调用方可通过 title 提示）
  return 'TEXT';
}

/** 按连接类型选择合适的归一化函数 */
export function normalizeDataTypeFor(connectionType: string | undefined, raw: string): string {
  if (connectionType === 'postgres') {
    return normalizeDataTypeForPG(raw);
  }
  return normalizeDataType(raw);
}
