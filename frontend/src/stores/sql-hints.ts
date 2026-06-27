import { create } from 'zustand';
import { schemaApi } from '@/lib/api/schema';
import type { ColumnDTO, TableDTO, ViewDTO } from '@/lib/api/types';

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
  tables: TableDTO[];
  views: ViewDTO[];
  columnsByTable: Record<string, ColumnDTO[]>;
  loading: boolean;
  error?: string;
  loadedAt?: number;
}

interface SqlHintState {
  catalogs: Record<string, SqlHintCatalog>;
  ensureCatalog: (ctx: SqlHintContext) => Promise<SqlHintCatalog | null>;
  ensureTableColumns: (ctx: SqlHintContext, table: string) => Promise<ColumnDTO[]>;
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
  const table = catalog.tables.find(t => tableKey(t.name) === target);
  if (table) return table.name;
  const view = catalog.views.find(v => tableKey(v.name) === target);
  return view?.name || null;
}

const catalogInflight: Partial<Record<string, Promise<SqlHintCatalog | null>>> = {};

export const useSqlHintStore = create<SqlHintState>((set, get) => ({
  catalogs: {},

  getCatalog: (ctx) => {
    const key = hintCatalogKey(ctx);
    return key ? get().catalogs[key] || null : null;
  },

  ensureCatalog: async (ctx) => {
    const normalized = normalizeHintContext(ctx);
    if (!normalized) return null;

    const key = hintCatalogKey(normalized)!;
    const existing = get().catalogs[key];
    if (existing?.loadedAt) {
      return existing;
    }
    const inflight = catalogInflight[key];
    if (inflight) {
      return inflight;
    }

    set(state => ({
      catalogs: {
        ...state.catalogs,
        [key]: {
          ...(state.catalogs[key] || {
            key,
            connID: normalized.connID,
            connectionType: normalized.connectionType,
            database: normalized.database,
            schema: normalized.schema,
            tables: [],
            views: [],
            columnsByTable: {},
          }),
          loading: true,
          error: undefined,
        },
      },
    }));

    catalogInflight[key] = (async () => {
      try {
        const [tables, views] = await Promise.all([
          schemaApi.listTables(normalized.connID, normalized.database, normalized.schema),
          schemaApi.listViews(normalized.connID, normalized.database, normalized.schema).catch(() => [] as ViewDTO[]),
        ]);
        const catalog: SqlHintCatalog = {
          key,
          connID: normalized.connID,
          connectionType: normalized.connectionType,
          database: normalized.database,
          schema: normalized.schema,
          tables: tables || [],
          views: views || [],
          columnsByTable: get().catalogs[key]?.columnsByTable || {},
          loading: false,
          loadedAt: Date.now(),
        };
        set(state => ({ catalogs: { ...state.catalogs, [key]: catalog } }));
        return catalog;
      } catch (e: any) {
        set(state => ({
          catalogs: {
            ...state.catalogs,
            [key]: {
              ...(state.catalogs[key] || {
                key,
                connID: normalized.connID,
                connectionType: normalized.connectionType,
                database: normalized.database,
                schema: normalized.schema,
                tables: [],
                views: [],
                columnsByTable: {},
              }),
              loading: false,
              error: e.message,
            },
          },
        }));
        return null;
      } finally {
        delete catalogInflight[key];
      }
    })();

    return catalogInflight[key] || null;
  },

  ensureTableColumns: async (ctx, table) => {
    const normalized = normalizeHintContext(ctx);
    if (!normalized || !table.trim()) return [];

    const catalog = await get().ensureCatalog(normalized);
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
    delete catalogInflight[key];
    set(state => {
      const catalogs = { ...state.catalogs };
      delete catalogs[key];
      return { catalogs };
    });
  },

  clearConnection: (connID) => set(state => {
    const catalogs: Record<string, SqlHintCatalog> = {};
    Object.entries(state.catalogs).forEach(([key, catalog]) => {
      if (catalog.connID !== connID) catalogs[key] = catalog;
      else delete catalogInflight[key];
    });
    return { catalogs };
  }),
}));
