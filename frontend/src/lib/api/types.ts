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
  rows: any[][];
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
