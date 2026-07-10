// 统一响应
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

// 分页结果
export interface PageResult<T> {
  page: number;
  pageSize: number;
  total: number;
  list: T[];
}

// 连接 DTO
export interface ConnectionDTO {
  id: number;
  name: string;
  type: string;
  host: string;
  port: number;
  user: string;
  database: string;
  ssl: boolean;
  group: string;
  color: string;
  lastConnectedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

// 连接请求
export interface CreateConnectionRequest {
  name: string;
  type: string;
  host: string;
  port: number;
  user: string;
  password: string;
  database: string;
  ssl: boolean;
  group: string;
  color: string;
}

export type UpdateConnectionRequest = CreateConnectionRequest;

export interface TestConnectionRequest {
  type: string;
  host: string;
  port: number;
  user: string;
  password: string;
  database: string;
  ssl: boolean;
}

export interface TestConnectionResultDTO {
  success: boolean;
  message: string;
  serverVersion: string;
}

export interface ListConnectionRequest {
  group?: string;
  type?: string;
  keyword?: string;
}

export interface PageConnectionRequest extends ListConnectionRequest {
  page?: number;
  pageSize?: number;
}

// Query DTOs
export interface QueryResultDTO {
  connID: number;
  sql: string;
  columns: ColumnDTO[];
  rows: (string | null)[][];
  rowsAffected: number;
  lastInsertID: number;
  durationMs: number;
  truncated: boolean;
}

export interface ColumnDTO {
  name: string;
  databaseType: string;
  nullable: boolean;
  isPrimary: boolean;
  autoIncrement: boolean;
  unsigned: boolean;
  defaultValue: string | null;
  comment: string;
  extra: string;
  ordinalPosition: number;
  typeLength: number | null;
  precision: number | null;
  scale: number | null;
  collation: string;
}

export interface ExecuteQueryRequest {
  connID: number;
  sql: string;
  maxRows?: number;
  database?: string;
  schema?: string;
}

export interface QueryHistoryRequest {
  connID?: number;
  keyword?: string;
  page?: number;
  pageSize?: number;
}

export interface QueryHistoryDTO {
  id: number;
  connID: number;
  sql: string;
  status: string;
  errorMessage: string;
  rowCount: number;
  durationMs: number;
  executedAt: string;
}

// Export DTOs
export interface ExportCSVRequest {
  filename: string;
  content: string;
}

export interface ExportCSVResult {
  saved: boolean;
  path: string;
}

// Schema DTOs
export interface DatabaseDTO {
  name: string;
  charset: string;
  collation: string;
}

export interface SchemaDTO {
  name: string;
  owner: string;
}

export interface TableDTO {
  name: string;
  type: string;
  comment: string;
  rowCount: number;
}

export interface ViewDTO {
  name: string;
  definition: string;
}

export interface TableDetailDTO {
  name: string;
  schema: string;
  database: string;
  comment: string;
  columns: ColumnDTO[];
  indexes: IndexDTO[];
  foreignKeys: ForeignKeyDTO[];
}

export interface IndexDTO {
  name: string;
  columns: string[];
  unique: boolean;
  primary: boolean;
  comment: string;
}

export interface ForeignKeyDTO {
  name: string;
  columns: string[];
  referencedSchema: string;
  referencedTable: string;
  referencedColumns: string[];
  onUpdate: string;
  onDelete: string;
}

export interface ListTablesRequest {
  connID: number;
  database: string;
  schema: string;
  limit: number;
  offset: number;
}

export interface SearchTablesRequest {
  connID: number;
  database: string;
  schema: string;
  keyword: string;
  limit: number;
}

// DDL DTOs
export interface CharsetDTO {
  name: string;
  description: string;
  defaultCollation: string;
}

export interface CollationDTO {
  name: string;
  charset: string;
  isDefault: boolean;
}

export interface ColumnSpecDTO {
  name: string;
  dataType: string;
  typeLength: number | null;
  precision: number | null;
  scale: number | null;
  nullable: boolean;
  autoIncrement: boolean;
  unsigned: boolean;
  defaultValue: string | null;
  comment: string;
  collation: string;
  first: boolean;
  after: string;
}

export interface IndexSpecDTO {
  name: string;
  type: string;
  columns: string[];
  comment: string;
}

export interface ForeignKeySpecDTO {
  name: string;
  columns: string[];
  referencedSchema: string;
  referencedTable: string;
  referencedColumns: string[];
  onDelete: string;
  onUpdate: string;
}

export interface ColumnRenameDTO {
  old: string;
  new: string;
}

export interface CreateDatabaseRequest {
  connID: number;
  name: string;
  charset: string;
  collation: string;
}

export interface DropDatabaseRequest {
  connID: number;
  name: string;
}

export interface CreateTableRequest {
  connID: number;
  database: string;
  schema: string;
  name: string;
  columns: ColumnSpecDTO[];
  primaryKey: string[];
  indexes: IndexSpecDTO[];
  foreignKeys: ForeignKeySpecDTO[];
  engine: string;
  charset: string;
  collation: string;
  comment: string;
}

export interface AlterTableRequest {
  connID: number;
  database: string;
  schema: string;
  name: string;
  addColumns: ColumnSpecDTO[];
  dropColumns: string[];
  renameColumns: ColumnRenameDTO[];
  modifyColumns: ColumnSpecDTO[];
  addIndexes: IndexSpecDTO[];
  modifyIndexes: IndexSpecDTO[];
  dropIndexes: string[];
  addForeignKeys: ForeignKeySpecDTO[];
  dropForeignKeys: string[];
  engine: string;
  charset: string;
  collation: string;
  comment: string;
}

export interface DropTableRequest {
  connID: number;
  database: string;
  schema: string;
  name: string;
}

// MCP Server DTOs
export interface McpServerStatusDTO {
  connID: number;
  enabled: boolean;
  address: string;
  port: number;
  token: string;
  allowSelect: boolean;
  allowInsert: boolean;
  allowUpdate: boolean;
  allowDelete: boolean;
  allowDDL: boolean;
}

export interface StartMcpServerRequest {
  connID: number;
  allowSelect: boolean;
  allowInsert: boolean;
  allowUpdate: boolean;
  allowDelete: boolean;
  allowDDL: boolean;
  forceNewPort?: boolean;
}

// SqlCompleteRequest SQL 补全请求
export interface SqlCompleteRequest {
  connID: number;
  connectionType?: string;
  database?: string;
  schema?: string;
  sql: string;
  cursorLine: number;
  cursorColumn: number;
  prefix?: string;
}

// SqlCompleteCandidate 单个补全候选
export interface SqlCompleteCandidate {
  label: string;
  kind: string;
  detail: string;
  insertText: string;
  sortText: string;
}
