import {
  ListDatabases, ListSchemas, ListTables, ListViews, DescribeTable,
  ListCharsets, ListCollations, GetCreateTableDDL, PreviewCreateTable,
  PreviewAlterTable, CreateDatabase, CreateTable, AlterTable,
  PreviewDropTable, DropTable,
} from '../../../wailsjs/go/http/SchemaAPI';
import type {
  ApiResponse, DatabaseDTO, SchemaDTO, TableDTO, ViewDTO, TableDetailDTO,
  CharsetDTO, CollationDTO, CreateTableRequest, AlterTableRequest, CreateDatabaseRequest,
  DropTableRequest,
} from './types';

function checkCode<T>(resp: ApiResponse<T>): T {
  if (resp.code !== 200) {
    throw new Error(resp.message || `Error code: ${resp.code}`);
  }
  return resp.data;
}

export const schemaApi = {
  listDatabases: (connID: number) =>
    ListDatabases(connID).then(r => checkCode(r as unknown as ApiResponse<DatabaseDTO[]>)),

  listSchemas: (connID: number, database: string) =>
    ListSchemas(connID, database).then(r => checkCode(r as unknown as ApiResponse<SchemaDTO[]>)),

  listTables: (connID: number, database: string, schema: string) =>
    ListTables(connID, database, schema).then(r => checkCode(r as unknown as ApiResponse<TableDTO[]>)),

  listViews: (connID: number, database: string, schema: string) =>
    ListViews(connID, database, schema).then(r => checkCode(r as unknown as ApiResponse<ViewDTO[]>)),

  describeTable: (connID: number, database: string, schema: string, table: string) =>
    DescribeTable(connID, database, schema, table).then(r => checkCode(r as unknown as ApiResponse<TableDetailDTO>)),

  listCharsets: (connID: number) =>
    ListCharsets({ connID } as any).then(r => checkCode(r as unknown as ApiResponse<CharsetDTO[]>)),

  listCollations: (connID: number, charset?: string) =>
    ListCollations({ connID, charset: charset ?? '' } as any).then(r => checkCode(r as unknown as ApiResponse<CollationDTO[]>)),

  getCreateTableDDL: (connID: number, database: string, schema: string, table: string) =>
    GetCreateTableDDL({ connID, database, schema, table } as any).then(r => checkCode(r as unknown as ApiResponse<string>)),

  previewCreateTable: (req: CreateTableRequest) =>
    PreviewCreateTable(req as any).then(r => checkCode(r as unknown as ApiResponse<string>)),

  previewAlterTable: (req: AlterTableRequest) =>
    PreviewAlterTable(req as any).then(r => checkCode(r as unknown as ApiResponse<string>)),

  createDatabase: (req: CreateDatabaseRequest) =>
    CreateDatabase(req as any).then(r => checkCode(r as unknown as ApiResponse<boolean>)),

  createTable: (req: CreateTableRequest) =>
    CreateTable(req as any).then(r => checkCode(r as unknown as ApiResponse<boolean>)),

  alterTable: (req: AlterTableRequest) =>
    AlterTable(req as any).then(r => checkCode(r as unknown as ApiResponse<boolean>)),

  previewDropTable: (req: DropTableRequest) =>
    PreviewDropTable(req as any).then(r => checkCode(r as unknown as ApiResponse<string>)),

  dropTable: (req: DropTableRequest) =>
    DropTable(req as any).then(r => checkCode(r as unknown as ApiResponse<boolean>)),
};
