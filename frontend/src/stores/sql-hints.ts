import { create } from 'zustand';
import { schemaApi } from '@/lib/api/schema';
import type { ColumnDTO, TableDTO } from '@/lib/api/types';

export interface SqlHintContext {
  connID: number;
  connectionType?: string;
  database?: string;
  schema?: string;
}

export interface SqlHintCatalog {
  key: string;
  connID: number;
  connectionType?: string;
  database: string;
  schema: string;
  columnsByTable: Record<string, ColumnDTO[]>;
}

interface SqlHintState {
  catalogs: Record<string, SqlHintCatalog>;
  searchCache: Record<string, TableDTO[]>;
  ensureCatalog: (ctx: SqlHintContext) => SqlHintCatalog | null;
  ensureTableColumns: (ctx: SqlHintContext, table: string) => Promise<ColumnDTO[]>;
  searchTables: (ctx: SqlHintContext, keyword: string) => Promise<TableDTO[]>;
  invalidateCatalog: (ctx: SqlHintContext) => void;
  clearConnection: (connID: number) => void;
  getCatalog: (ctx: SqlHintContext) => SqlHintCatalog | null;
}

function normalizeSchema(ctx: SqlHintContext): string {
  const schema = (ctx.schema || '').trim();
  if (schema) return schema;
  if (ctx.connectionType === 'postgres') return 'public';
  if (ctx.connectionType === 'sqlite') return 'main';
  return (ctx.database || '').trim();
}

export function normalizeHintContext(ctx: SqlHintContext): (SqlHintContext & { database: string; schema: string }) | null {
  const database = (ctx.database || '').trim();
  if (!ctx.connID || !database) return null;
  return {
    ...ctx,
    database,
    schema: normalizeSchema(ctx),
  };
}

export function hintCatalogKey(ctx: SqlHintContext): string | null {
  const normalized = normalizeHintContext(ctx);
  if (!normalized) return null;
  return `${normalized.connID}:${normalized.database}:${normalized.schema}`;
}

function tableKey(name: string): string {
  return name.trim().toLowerCase();
}

function findRelationName(catalog: SqlHintCatalog, name: string): string | null {
  const target = tableKey(name);
  const cached = Object.keys(catalog.columnsByTable).find(k => tableKey(k) === target);
  return cached || null;
}

export const useSqlHintStore = create<SqlHintState>((set, get) => ({
  catalogs: {},
  searchCache: {},

  getCatalog: (ctx) => {
    const key = hintCatalogKey(ctx);
    return key ? get().catalogs[key] || null : null;
  },

  ensureCatalog: (ctx) => {
    const normalized = normalizeHintContext(ctx);
    if (!normalized) return null;
    const key = hintCatalogKey(normalized)!;
    const existing = get().catalogs[key];
    if (existing) return existing;

    const catalog: SqlHintCatalog = {
      key,
      connID: normalized.connID,
      connectionType: normalized.connectionType,
      database: normalized.database,
      schema: normalized.schema,
      columnsByTable: {},
    };
    set(state => ({ catalogs: { ...state.catalogs, [key]: catalog } }));
    return catalog;
  },

  searchTables: async (ctx, keyword) => {
    const normalized = normalizeHintContext(ctx);
    if (!normalized || !keyword.trim()) return [];

    const cacheKey = `${normalized.connID}:${normalized.database}:${normalized.schema}:${keyword.trim().toLowerCase()}`;
    const cached = get().searchCache[cacheKey];
    if (cached) return cached;

    try {
      const list = await schemaApi.searchTables({
        connID: normalized.connID,
        database: normalized.database,
        schema: normalized.schema,
        keyword: keyword.trim(),
        limit: 50,
      });
      set(state => ({ searchCache: { ...state.searchCache, [cacheKey]: list || [] } }));
      return list || [];
    } catch {
      return [];
    }
  },

  ensureTableColumns: async (ctx, table) => {
    const normalized = normalizeHintContext(ctx);
    if (!normalized || !table.trim()) return [];

    const catalog = get().ensureCatalog(normalized);
    if (!catalog) return [];

    const relationName = findRelationName(catalog, table) || table.trim();
    const columnKey = tableKey(relationName);
    const current = get().catalogs[catalog.key];
    if (current?.columnsByTable[columnKey]) {
      return current.columnsByTable[columnKey];
    }

    try {
      const detail = await schemaApi.describeTable(
        normalized.connID,
        normalized.database,
        normalized.schema,
        relationName,
      );
      const columns = detail?.columns || [];
      set(state => {
        const latest = state.catalogs[catalog.key] || catalog;
        return {
          catalogs: {
            ...state.catalogs,
            [catalog.key]: {
              ...latest,
              columnsByTable: {
                ...latest.columnsByTable,
                [columnKey]: columns,
              },
            },
          },
        };
      });
      return columns;
    } catch {
      return [];
    }
  },

  invalidateCatalog: (ctx) => {
    const key = hintCatalogKey(ctx);
    if (!key) return;
    set(state => {
      const catalogs = { ...state.catalogs };
      delete catalogs[key];
      const searchCache: Record<string, TableDTO[]> = {};
      Object.entries(state.searchCache).forEach(([k, v]) => {
        if (!k.startsWith(`${ctx.connID}:`)) searchCache[k] = v;
      });
      return { catalogs, searchCache };
    });
  },

  clearConnection: (connID) => set(state => {
    const catalogs: Record<string, SqlHintCatalog> = {};
    Object.entries(state.catalogs).forEach(([key, catalog]) => {
      if (catalog.connID !== connID) catalogs[key] = catalog;
    });
    const searchCache: Record<string, TableDTO[]> = {};
    Object.entries(state.searchCache).forEach(([key, tables]) => {
      if (!key.startsWith(`${connID}:`)) searchCache[key] = tables;
    });
    return { catalogs, searchCache };
  }),
}));
