import { create } from 'zustand';
import type { QueryResultDTO, ExecuteQueryRequest, QueryHistoryDTO } from '@/lib/api/types';
import { queryApi } from '@/lib/api/query';
import { toast } from '@/components/ui/use-toast';

interface QueryState {
  // 当前活跃查询结果（按 tab 索引）
  results: QueryResultDTO[];
  activeResultIndex: number;
  executing: boolean;
  history: QueryHistoryDTO[];
  historyLoading: boolean;

  // 操作
  execute: (req: ExecuteQueryRequest) => Promise<QueryResultDTO | null>;
  cancel: (connID: number) => Promise<void>;
  setActiveResult: (index: number) => void;
  addEmptyResult: () => void;
  closeResult: (index: number) => void;
  closeOtherResults: (keepIndex: number) => void;
  closeAllResults: () => void;
  fetchHistory: (connID: number) => Promise<void>;
}

export const useQueryStore = create<QueryState>((set, get) => ({
  results: [],
  activeResultIndex: 0,
  executing: false,
  history: [],
  historyLoading: false,

  execute: async (req) => {
    set({ executing: true });
    try {
      const result = await queryApi.execute(req);
      set(state => {
        const results = [...state.results, result];
        return { results, activeResultIndex: results.length - 1, executing: false };
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

  setActiveResult: (index) => set({ activeResultIndex: index }),

  addEmptyResult: () => set(state => ({
    results: [...state.results, { connID: 0, sql: '', columns: [], rows: [], rowsAffected: 0, lastInsertID: 0, durationMs: 0, truncated: false }],
    activeResultIndex: state.results.length,
  })),

  closeResult: (index) => set(state => {
    const results = state.results.filter((_, i) => i !== index);
    const activeResultIndex = Math.min(state.activeResultIndex, Math.max(0, results.length - 1));
    return { results, activeResultIndex };
  }),

  closeOtherResults: (keepIndex) => set(state => {
    const results = [state.results[keepIndex]].filter(Boolean);
    return { results, activeResultIndex: 0 };
  }),

  closeAllResults: () => set({ results: [], activeResultIndex: 0 }),

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
