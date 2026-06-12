import { create } from 'zustand';
import type { TableDetailDTO } from '@/lib/api/types';
import { schemaApi } from '@/lib/api/schema';
import { toast } from '@/components/ui/use-toast';

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
}

interface SchemaState {
  // 按 connID 缓存的树数据
  trees: Record<number, TreeNode[]>;
  expandedKeys: Set<string>;
  selectedNode: TreeNode | null;
  selectedTable: TableDetailDTO | null;

  // 操作
  setSelectedNode: (node: TreeNode | null) => void;
  loadDatabases: (connID: number) => Promise<void>;
  loadTables: (connID: number, database: string, schema: string) => Promise<void>;
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

  loadTables: async (connID, database, schema) => {
    try {
      const [tables, views] = await Promise.all([
        schemaApi.listTables(connID, database, schema),
        schemaApi.listViews(connID, database, schema).catch(() => [] as any[]),
      ]);

      const tableNodes: TreeNode[] = tables.map(t => ({
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

      const viewNodes: TreeNode[] = views.map((v: any) => ({
        key: `conn:${connID}:db:${database}:schema:${schema}:view:${v.name}`,
        label: v.name,
        type: 'view' as const,
        connID,
        database,
        schema,
        tableName: v.name,
        loaded: true, // views don't have sub-children in V1.0
      }));

      const allNodes = [...tableNodes, ...viewNodes];

      set(state => {
        const trees = { ...state.trees };
        const dbNodes = trees[connID] || [];
        const updateChildren = (nodes: TreeNode[]): TreeNode[] =>
          nodes.map(n => {
            if (n.key === `conn:${connID}:db:${database}`) {
              return { ...n, children: allNodes, loaded: true };
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
    // 重新加载该数据库下的表和视图
    // MySQL 的 schema 默认与 database 同名
    await get().loadTables(connID, db, db);
  },

  invalidateTable: async (connID, db, schema, table) => {
    // 重新加载表详情
    await get().loadTableDetail(connID, db, schema, table);
  },

  refreshTree: async (connID) => {
    // 刷新指定连接的整个树
    get().resetTree(connID);
    await get().loadDatabases(connID);
  },

  getConnIDForNode: (node) => {
    return node.connID;
  },
}));
