import { create } from 'zustand';
import type { TableDetailDTO, TableDTO } from '@/lib/api/types';
import { schemaApi } from '@/lib/api/schema';
import { toast } from '@/components/ui/use-toast';

// 每批加载的表/视图数量
export const DEFAULT_TABLE_PAGE_SIZE = 200;
export const DEFAULT_SEARCH_LIMIT = 50;

// 树节点
export interface TreeNode {
  key: string;       // 唯一 key，如 "conn:1:db:mysql:table:users"
  label: string;
  type: 'connection' | 'database' | 'schema' | 'table' | 'view' | 'column' | 'index' | 'foreignKey' | 'columns-folder' | 'indexes-folder' | 'foreign-keys-folder';
  icon?: string;
  children?: TreeNode[];
  loaded?: boolean;  // 子节点是否已懒加载
  connID: number;
  database?: string;
  schema?: string;
  tableName?: string;
  columnName?: string;  // 字段节点：字段名（label 含类型后缀，此字段仅名称）
  isPrimary?: boolean;  // 字段节点：是否为主键
  // 分页状态（仅 database/schema 节点使用）
  tableOffset?: number;
  hasMoreTables?: boolean;
  viewOffset?: number;
  hasMoreViews?: boolean;
}

// 搜索结果状态
interface SearchState {
  keyword: string;
  results: TreeNode[];
  loading: boolean;
}

interface SchemaState {
  // 按 connID 缓存的树数据
  trees: Record<number, TreeNode[]>;
  expandedKeys: Set<string>;
  selectedNode: TreeNode | null;
  selectedTable: TableDetailDTO | null;
  // 按节点 key 缓存的搜索结果
  searchMap: Record<string, SearchState>;

  // 操作
  setSelectedNode: (node: TreeNode | null) => void;
  loadDatabases: (connID: number) => Promise<void>;
  loadSchemas: (connID: number, database: string) => Promise<void>;
  /** 加载/刷新指定 schema/database 下的第一批表和视图 */
  loadTables: (connID: number, database: string, schema: string) => Promise<void>;
  /** 加载更多表 */
  loadMoreTables: (nodeKey: string) => Promise<void>;
  /** 加载更多视图 */
  loadMoreViews: (nodeKey: string) => Promise<void>;
  /** 按关键字搜索表/视图 */
  searchTables: (connID: number, database: string, schema: string, keyword: string) => Promise<TreeNode[]>;
  loadTableDetail: (connID: number, database: string, schema: string, table: string) => Promise<void>;
  toggleExpand: (key: string) => void;
  resetTree: (connID: number) => void;

  // 缓存失效
  invalidateDatabase: (connID: number, db: string) => Promise<void>;
  invalidateTable: (connID: number, db: string, schema: string, table: string) => Promise<void>;

  // 便利方法
  refreshTree: (connID: number) => Promise<void>;
  getConnIDForNode: (node: TreeNode) => number;
}

export const useSchemaStore = create<SchemaState>((set, get) => ({
  trees: {},
  expandedKeys: new Set(),
  selectedNode: null,
  selectedTable: null,
  searchMap: {},

  setSelectedNode: (node) => set({ selectedNode: node }),

  loadDatabases: async (connID) => {
    try {
      const dbs = await schemaApi.listDatabases(connID);
      const nodes: TreeNode[] = dbs.map(db => ({
        key: `conn:${connID}:db:${db.name}`,
        label: db.name,
        type: 'database',
        connID,
        database: db.name,
        loaded: false,
        children: [],
      }));
      set(state => ({ trees: { ...state.trees, [connID]: nodes } }));
    } catch (e: any) {
      toast({ title: '加载数据库列表失败', description: e.message, variant: 'destructive' });
    }
  },

  loadSchemas: async (connID, database) => {
    try {
      const schemas = await schemaApi.listSchemas(connID, database);
      const nodes: TreeNode[] = schemas.map(s => ({
        key: `conn:${connID}:db:${database}:schema:${s.name}`,
        label: s.name,
        type: 'schema' as const,
        connID,
        database,
        schema: s.name,
        loaded: false,
        children: [],
      }));

      set(state => {
        const trees = { ...state.trees };
        const dbNodes = trees[connID] || [];
        const updateChildren = (items: TreeNode[]): TreeNode[] =>
          items.map(n => {
            if (n.key === `conn:${connID}:db:${database}`) {
              return { ...n, children: nodes, loaded: true };
            }
            if (n.children) {
              return { ...n, children: updateChildren(n.children) };
            }
            return n;
          });
        trees[connID] = updateChildren(dbNodes);
        return { trees };
      });
    } catch (e: any) {
      toast({ title: '加载 Schema 列表失败', description: e.message, variant: 'destructive' });
    }
  },

  loadTables: async (connID, database, schema) => {
    try {
      const targetSchema = schema || database;
      const [tables, views] = await Promise.all([
        schemaApi.listTables(connID, database, targetSchema, DEFAULT_TABLE_PAGE_SIZE, 0),
        schemaApi.listViews(connID, database, targetSchema, DEFAULT_TABLE_PAGE_SIZE, 0).catch(() => [] as TableDTO[]),
      ]);

      const tableNodes: TreeNode[] = tables.map(t => ({
        key: `conn:${connID}:db:${database}:schema:${targetSchema}:table:${t.name}`,
        label: t.name,
        type: 'table' as const,
        connID,
        database,
        schema: targetSchema,
        tableName: t.name,
        loaded: false,
        children: [],
      }));

      const viewNodes: TreeNode[] = views.map((v: any) => ({
        key: `conn:${connID}:db:${database}:schema:${targetSchema}:view:${v.name}`,
        label: v.name,
        type: 'view' as const,
        connID,
        database,
        schema: targetSchema,
        tableName: v.name,
        loaded: true,
      }));

      const allNodes = [...tableNodes, ...viewNodes];

      set(state => {
        const trees = { ...state.trees };
        const dbNodes = trees[connID] || [];
        const updateChildren = (nodes: TreeNode[]): TreeNode[] =>
          nodes.map(n => {
            if (
              n.key === `conn:${connID}:db:${database}:schema:${targetSchema}` ||
              (n.key === `conn:${connID}:db:${database}` && n.children?.every(child => child.type !== 'schema'))
            ) {
              return {
                ...n,
                children: allNodes,
                loaded: true,
                tableOffset: tableNodes.length,
                hasMoreTables: tables.length >= DEFAULT_TABLE_PAGE_SIZE,
                viewOffset: viewNodes.length,
                hasMoreViews: views.length >= DEFAULT_TABLE_PAGE_SIZE,
              };
            }
            if (n.children) {
              return { ...n, children: updateChildren(n.children) };
            }
            return n;
          });
        trees[connID] = updateChildren(dbNodes);
        return { trees };
      });
    } catch (e: any) {
      toast({ title: '加载表列表失败', description: e.message, variant: 'destructive' });
    }
  },

  loadMoreTables: async (nodeKey) => {
    const node = findNode(get().trees, nodeKey);
    if (!node || !node.database) return;
    if (node.hasMoreTables === false) return;

    const connID = node.connID;
    const database = node.database;
    const schema = node.schema || database;
    const offset = node.tableOffset || 0;

    try {
      const tables = await schemaApi.listTables(connID, database, schema, DEFAULT_TABLE_PAGE_SIZE, offset);
      const newNodes: TreeNode[] = tables.map(t => ({
        key: `conn:${connID}:db:${database}:schema:${schema}:table:${t.name}`,
        label: t.name,
        type: 'table' as const,
        connID,
        database,
        schema,
        tableName: t.name,
        loaded: false,
        children: [],
      }));

      set(state => {
        const trees = { ...state.trees };
        const updateChildren = (nodes: TreeNode[]): TreeNode[] =>
          nodes.map(n => {
            if (n.key === nodeKey) {
              const existing = n.children || [];
              // 视图节点放在表节点后面；这里简单追加到末尾
              return {
                ...n,
                children: [...existing, ...newNodes],
                tableOffset: offset + newNodes.length,
                hasMoreTables: tables.length >= DEFAULT_TABLE_PAGE_SIZE,
              };
            }
            if (n.children) {
              return { ...n, children: updateChildren(n.children) };
            }
            return n;
          });
        trees[connID] = updateChildren(trees[connID] || []);
        return { trees };
      });
    } catch (e: any) {
      toast({ title: '加载更多表失败', description: e.message, variant: 'destructive' });
    }
  },

  loadMoreViews: async (nodeKey) => {
    const node = findNode(get().trees, nodeKey);
    if (!node || !node.database) return;
    if (node.hasMoreViews === false) return;

    const connID = node.connID;
    const database = node.database;
    const schema = node.schema || database;
    const offset = node.viewOffset || 0;

    try {
      const views = await schemaApi.listViews(connID, database, schema, DEFAULT_TABLE_PAGE_SIZE, offset);
      const newNodes: TreeNode[] = views.map((v: any) => ({
        key: `conn:${connID}:db:${database}:schema:${schema}:view:${v.name}`,
        label: v.name,
        type: 'view' as const,
        connID,
        database,
        schema,
        tableName: v.name,
        loaded: true,
      }));

      set(state => {
        const trees = { ...state.trees };
        const updateChildren = (nodes: TreeNode[]): TreeNode[] =>
          nodes.map(n => {
            if (n.key === nodeKey) {
              const existing = n.children || [];
              return {
                ...n,
                children: [...existing, ...newNodes],
                viewOffset: offset + newNodes.length,
                hasMoreViews: views.length >= DEFAULT_TABLE_PAGE_SIZE,
              };
            }
            if (n.children) {
              return { ...n, children: updateChildren(n.children) };
            }
            return n;
          });
        trees[connID] = updateChildren(trees[connID] || []);
        return { trees };
      });
    } catch (e: any) {
      toast({ title: '加载更多视图失败', description: e.message, variant: 'destructive' });
    }
  },

  searchTables: async (connID, database, schema, keyword) => {
    const targetSchema = schema || database;
    const nodeKey = schema
      ? `conn:${connID}:db:${database}:schema:${targetSchema}`
      : `conn:${connID}:db:${database}`;
    const trimmed = keyword.trim().toLowerCase();

    set(state => ({
      searchMap: {
        ...state.searchMap,
        [nodeKey]: { keyword: trimmed, results: [], loading: true },
      },
    }));

    try {
      const list = await schemaApi.searchTables({
        connID,
        database,
        schema: targetSchema,
        keyword: trimmed,
        limit: DEFAULT_SEARCH_LIMIT,
      });
      const nodes: TreeNode[] = list.map(t => ({
        key: `conn:${connID}:db:${database}:schema:${targetSchema}:${t.type === 'VIEW' || t.type === 'view' ? 'view' : 'table'}:${t.name}`,
        label: t.name,
        type: (t.type === 'VIEW' || t.type === 'view' ? 'view' : 'table') as 'table' | 'view',
        connID,
        database,
        schema: targetSchema,
        tableName: t.name,
        loaded: t.type === 'VIEW' || t.type === 'view',
      }));
      set(state => ({
        searchMap: {
          ...state.searchMap,
          [nodeKey]: { keyword: trimmed, results: nodes, loading: false },
        },
      }));
      return nodes;
    } catch (e: any) {
      set(state => ({
        searchMap: {
          ...state.searchMap,
          [nodeKey]: { keyword: trimmed, results: [], loading: false },
        },
      }));
      toast({ title: '搜索表失败', description: e.message, variant: 'destructive' });
      return [];
    }
  },

  loadTableDetail: async (connID, database, schema, table) => {
    try {
      const detail = await schemaApi.describeTable(connID, database, schema, table);
      set({ selectedTable: detail });

      // 同时更新树节点的 children（使用文件夹层级结构）
      set(state => {
        const trees = { ...state.trees };
        const updateChildren = (nodes: TreeNode[]): TreeNode[] =>
          nodes.map(n => {
            if (n.key === `conn:${connID}:db:${database}:schema:${schema}:table:${table}`) {
              const colNodes: TreeNode[] = (detail.columns ?? []).map(c => ({
                key: `${n.key}:folder:columns:col:${c.name}`,
                label: `${c.name} (${c.databaseType})`,
                type: 'column' as const,
                connID,
                database,
                schema,
                tableName: table,
                columnName: c.name,
                isPrimary: !!c.isPrimary,
                loaded: true,
              }));

              const idxNodes: TreeNode[] = (detail.indexes ?? []).map(idx => ({
                key: `${n.key}:folder:indexes:idx:${idx.name}`,
                label: idx.name,
                type: 'index' as const,
                connID,
                database,
                schema,
                tableName: table,
                loaded: true,
              }));

              const fkNodes: TreeNode[] = (detail.foreignKeys ?? []).map(fk => ({
                key: `${n.key}:folder:foreign-keys:fk:${fk.name}`,
                label: fk.name,
                type: 'foreignKey' as const,
                connID,
                database,
                schema,
                tableName: table,
                loaded: true,
              }));

              const folderChildren: TreeNode[] = [
                {
                  key: `${n.key}:folder:columns`,
                  label: 'Columns',
                  type: 'columns-folder' as const,
                  connID,
                  database,
                  schema,
                  tableName: table,
                  loaded: true,
                  children: colNodes,
                },
                {
                  key: `${n.key}:folder:indexes`,
                  label: 'Indexes',
                  type: 'indexes-folder' as const,
                  connID,
                  database,
                  schema,
                  tableName: table,
                  loaded: true,
                  children: idxNodes,
                },
                {
                  key: `${n.key}:folder:foreign-keys`,
                  label: 'Foreign Keys',
                  type: 'foreign-keys-folder' as const,
                  connID,
                  database,
                  schema,
                  tableName: table,
                  loaded: true,
                  children: fkNodes,
                },
              ];

              return { ...n, children: folderChildren, loaded: true };
            }
            if (n.children) {
              return { ...n, children: updateChildren(n.children) };
            }
            return n;
          });
        trees[connID] = updateChildren(trees[connID] || []);
        return { trees };
      });
    } catch (e: any) {
      toast({ title: '加载表结构失败', description: e.message, variant: 'destructive' });
    }
  },

  toggleExpand: (key) => set(state => {
    const expanded = new Set(state.expandedKeys);
    if (expanded.has(key)) expanded.delete(key); else expanded.add(key);
    return { expandedKeys: expanded };
  }),

  resetTree: (connID) => set(state => {
    const trees = { ...state.trees };
    delete trees[connID];
    const selectedNode = state.selectedNode?.connID === connID ? null : state.selectedNode;
    const selectedTable = state.selectedNode?.connID === connID ? null : state.selectedTable;
    return { trees, selectedNode, selectedTable };
  }),

  invalidateDatabase: async (connID, db) => {
    await get().loadTables(connID, db, db);
  },

  invalidateTable: async (connID, db, schema, table) => {
    await get().loadTableDetail(connID, db, schema, table);
  },

  refreshTree: async (connID) => {
    get().resetTree(connID);
    await get().loadDatabases(connID);
  },

  getConnIDForNode: (node) => {
    return node.connID;
  },
}));

// 在树中查找指定 key 的节点
function findNode(trees: Record<number, TreeNode[]>, key: string): TreeNode | null {
  for (const connID of Object.keys(trees).map(Number)) {
    const found = findInNodes(trees[connID], key);
    if (found) return found;
  }
  return null;
}

function findInNodes(nodes: TreeNode[] | undefined, key: string): TreeNode | null {
  if (!nodes) return null;
  for (const n of nodes) {
    if (n.key === key) return n;
    const found = findInNodes(n.children, key);
    if (found) return found;
  }
  return null;
}
