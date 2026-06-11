import { ListDatabases, ListSchemas, ListTables, ListViews, DescribeTable } from '../../../wailsjs/go/http/SchemaAPI';
import type { ApiResponse, DatabaseDTO, SchemaDTO, TableDTO, ViewDTO, TableDetailDTO } from './types';

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
};
