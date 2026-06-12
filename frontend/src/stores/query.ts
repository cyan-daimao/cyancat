import { create } from 'zustand';
import type { QueryResultDTO, ExecuteQueryRequest, QueryHistoryDTO } from '@/lib/api/types';
import { queryApi } from '@/lib/api/query';
import { toast } from '@/components/ui/use-toast';

export interface QueryTab {
  id: string;
  title: string;
  sql: string;
  connID: number;
  connectionType?: string;
  database?: string;
  schema?: string;
  contextLabel?: string;
  results: QueryResultDTO[];
  activeResultIndex: number;
}

export interface QueryTabContext {
  connID: number;
  connectionType?: string;
  database?: string;
  schema?: string;
  contextLabel?: string;
}

let nextQueryTabId = 1;

function createQueryTab(title?: string, context?: QueryTabContext): QueryTab {
  const id = `query-${nextQueryTabId++}`;
  return {
    id,
    title: title || `查询 ${nextQueryTabId - 1}`,
    sql: 'SELECT 1;',
    connID: context?.connID || 0,
    connectionType: context?.connectionType,
    database: context?.database,
    schema: context?.schema,
    contextLabel: context?.contextLabel,
    results: [],
    activeResultIndex: 0,
  };
}

interface QueryState {
  // SQL 编辑器标签页
  queryTabs: QueryTab[];
  activeQueryTabId: string;

  // 当前活跃查询结果（兼容旧调用，实际以 active query tab 为准）
  results: QueryResultDTO[];
  activeResultIndex: number;
  executing: boolean;
  history: QueryHistoryDTO[];
  historyLoading: boolean;

  // 操作
  createQueryTab: (context?: QueryTabContext) => void;
  closeQueryTab: (id: string) => void;
  closeOtherQueryTabs: (id: string) => void;
  closeQueryTabsToRight: (id: string) => void;
  closeQueryTabsToLeft: (id: string) => void;
  setActiveQueryTab: (id: string) => void;
  updateActiveTabSql: (sql: string) => void;
  updateActiveTabContext: (context: QueryTabContext) => void;
  execute: (req: ExecuteQueryRequest) => Promise<QueryResultDTO | null>;
  cancel: (connID: number) => Promise<void>;
  setActiveResult: (index: number) => void;
  addEmptyResult: (context?: QueryTabContext) => void;
  closeResult: (index: number) => void;
  closeOtherResults: (keepIndex: number) => void;
  closeAllResults: () => void;
  fetchHistory: (connID: number) => Promise<void>;
}

export const useQueryStore = create<QueryState>((set, get) => ({
  queryTabs: [createQueryTab()],
  activeQueryTabId: 'query-1',
  results: [],
  activeResultIndex: 0,
  executing: false,
  history: [],
  historyLoading: false,

  createQueryTab: (context) => set(state => {
    const tab = createQueryTab(undefined, context);
    return {
      queryTabs: [...state.queryTabs, tab],
      activeQueryTabId: tab.id,
      results: tab.results,
      activeResultIndex: tab.activeResultIndex,
    };
  }),

  closeQueryTab: (id) => set(state => {
    const currentIndex = state.queryTabs.findIndex(t => t.id === id);
    let queryTabs = state.queryTabs.filter(t => t.id !== id);
    if (queryTabs.length === 0) {
      queryTabs = [createQueryTab()];
    }

    let activeQueryTabId = state.activeQueryTabId;
    if (id === state.activeQueryTabId) {
      const nextIndex = Math.min(Math.max(currentIndex, 0), queryTabs.length - 1);
      activeQueryTabId = queryTabs[nextIndex].id;
    }

    const active = queryTabs.find(t => t.id === activeQueryTabId) || queryTabs[0];
    return {
      queryTabs,
      activeQueryTabId: active.id,
      results: active.results,
      activeResultIndex: active.activeResultIndex,
    };
  }),

  closeOtherQueryTabs: (id) => set(state => {
    const target = state.queryTabs.find(tab => tab.id === id);
    if (!target) return {};

    return {
      queryTabs: [target],
      activeQueryTabId: target.id,
      results: target.results,
      activeResultIndex: target.activeResultIndex,
    };
  }),

  closeQueryTabsToRight: (id) => set(state => {
    const targetIndex = state.queryTabs.findIndex(tab => tab.id === id);
    if (targetIndex < 0) return {};

    const queryTabs = state.queryTabs.slice(0, targetIndex + 1);
    const activeStillOpen = queryTabs.some(tab => tab.id === state.activeQueryTabId);
    const active = activeStillOpen
      ? queryTabs.find(tab => tab.id === state.activeQueryTabId)!
      : queryTabs[targetIndex];

    return {
      queryTabs,
      activeQueryTabId: active.id,
      results: active.results,
      activeResultIndex: active.activeResultIndex,
    };
  }),

  closeQueryTabsToLeft: (id) => set(state => {
    const targetIndex = state.queryTabs.findIndex(tab => tab.id === id);
    if (targetIndex < 0) return {};

    const queryTabs = state.queryTabs.slice(targetIndex);
    const activeStillOpen = queryTabs.some(tab => tab.id === state.activeQueryTabId);
    const active = activeStillOpen
      ? queryTabs.find(tab => tab.id === state.activeQueryTabId)!
      : queryTabs[0];

    return {
      queryTabs,
      activeQueryTabId: active.id,
      results: active.results,
      activeResultIndex: active.activeResultIndex,
    };
  }),

  setActiveQueryTab: (id) => set(state => {
    const active = state.queryTabs.find(t => t.id === id);
    if (!active) return {};
    return {
      activeQueryTabId: id,
      results: active.results,
      activeResultIndex: active.activeResultIndex,
    };
  }),

  updateActiveTabSql: (sql) => set(state => ({
    queryTabs: state.queryTabs.map(tab =>
      tab.id === state.activeQueryTabId ? { ...tab, sql } : tab
    ),
  })),

  updateActiveTabContext: (context) => set(state => ({
    queryTabs: state.queryTabs.map(tab =>
      tab.id === state.activeQueryTabId
        ? {
            ...tab,
            connID: context.connID,
            connectionType: context.connectionType,
            database: context.database,
            schema: context.schema,
            contextLabel: context.contextLabel,
          }
        : tab
    ),
  })),

  execute: async (req) => {
    set({ executing: true });
    try {
      const result = await queryApi.execute(req);
      set(state => {
        const queryTabs = state.queryTabs.map(tab => {
          if (tab.id !== state.activeQueryTabId) return tab;
          const results = [...tab.results, result];
          return { ...tab, results, activeResultIndex: results.length - 1 };
        });
        const active = queryTabs.find(tab => tab.id === state.activeQueryTabId) || queryTabs[0];
        return {
          queryTabs,
          results: active?.results || [],
          activeResultIndex: active?.activeResultIndex || 0,
          executing: false,
        };
      });
      return result;
    } catch (e: any) {
      toast({ title: '执行失败', description: e.message, variant: 'destructive' });
      set({ executing: false });
      return null;
    }
  },

  cancel: async (connID) => {
    await queryApi.cancel(connID);
    set({ executing: false });
  },

  setActiveResult: (index) => set(state => ({
    queryTabs: state.queryTabs.map(tab =>
      tab.id === state.activeQueryTabId ? { ...tab, activeResultIndex: index } : tab
    ),
    activeResultIndex: index,
  })),

  addEmptyResult: (context) => get().createQueryTab(context),

  closeResult: (index) => set(state => {
    const queryTabs = state.queryTabs.map(tab => {
      if (tab.id !== state.activeQueryTabId) return tab;
      const results = tab.results.filter((_, i) => i !== index);
      const activeResultIndex = Math.min(tab.activeResultIndex, Math.max(0, results.length - 1));
      return { ...tab, results, activeResultIndex };
    });
    const active = queryTabs.find(tab => tab.id === state.activeQueryTabId) || queryTabs[0];
    return {
      queryTabs,
      results: active?.results || [],
      activeResultIndex: active?.activeResultIndex || 0,
    };
  }),

  closeOtherResults: (keepIndex) => set(state => {
    const queryTabs = state.queryTabs.map(tab => {
      if (tab.id !== state.activeQueryTabId) return tab;
      const results = [tab.results[keepIndex]].filter(Boolean);
      return { ...tab, results, activeResultIndex: 0 };
    });
    const active = queryTabs.find(tab => tab.id === state.activeQueryTabId) || queryTabs[0];
    return { queryTabs, results: active?.results || [], activeResultIndex: 0 };
  }),

  closeAllResults: () => set(state => ({
    queryTabs: state.queryTabs.map(tab =>
      tab.id === state.activeQueryTabId ? { ...tab, results: [], activeResultIndex: 0 } : tab
    ),
    results: [],
    activeResultIndex: 0,
  })),

  fetchHistory: async (connID) => {
    set({ historyLoading: true });
    try {
      const page = await queryApi.history({ connID, pageSize: 50 });
      set({ history: page.list || [], historyLoading: false });
    } catch {
      set({ historyLoading: false });
    }
  },
}));
