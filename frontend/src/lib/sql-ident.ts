// SQL 标识符的方言感知引号工具，供 SQL 编辑器自动补全与对象树右键菜单共用。
// 集中在此以避免多处分散实现导致行为不一致（曾出现 postgres 上误用反引号）。

export type Dialect = 'mysql' | 'postgres' | 'sqlite';

/**
 * 把连接类型字符串归一化为方言枚举。
 * 容错识别常见变体（postgresql / pg / mariadb），避免精确匹配 'postgres' 漏掉旧数据。
 * 兜底返回 'mysql'。
 */
export function resolveDialect(type?: string): Dialect {
  const t = (type || '').trim().toLowerCase();
  if (/postgres|postgresql|pg/.test(t)) return 'postgres';
  if (/sqlite/.test(t)) return 'sqlite';
  if (/mysql|maria/.test(t)) return 'mysql';
  return 'mysql';
}

/** 简单标识符（字母/下划线/数字，且不以数字开头）无需引号。 */
function isSimpleName(name: string): boolean {
  return /^[A-Za-z_][A-Za-z0-9_]*$/.test(name);
}

/**
 * 按方言给标识符加引号。
 * - 简单名直接返回（多数情况，可读性更好）。
 * - postgres/sqlite 用双引号，内部 `"` 翻倍。
 * - mysql 用反引号，内部 `` ` `` 翻倍。
 */
export function quoteIdent(name: string, dialect: Dialect): string {
  if (!name) return name;
  if (isSimpleName(name)) return name;
  if (dialect === 'postgres' || dialect === 'sqlite') {
    return `"${name.split('"').join('""')}"`;
  }
  return `\`${name.split('`').join('``')}\``;
}

/**
 * 构造限定表名（schema.table / database.table）。
 * - postgres：`"schema"."table"`，schema 缺省 `public`。
 * - sqlite：`"main"."table"`。
 * - mysql：`` `database`.`table` ``。
 */
export function qualifiedTableName(opts: {
  dialect: Dialect;
  database?: string;
  schema?: string;
  table: string;
}): string {
  const { dialect, database, schema, table } = opts;
  if (dialect === 'postgres') {
    const s = schema || 'public';
    return `${quoteIdent(s, dialect)}.${quoteIdent(table, dialect)}`;
  }
  if (dialect === 'sqlite') {
    return `${quoteIdent('main', dialect)}.${quoteIdent(table, dialect)}`;
  }
  return `${quoteIdent(database || '', dialect)}.${quoteIdent(table, dialect)}`;
}
