import { create } from 'zustand';
import type { ConnectionDTO, CreateConnectionRequest, UpdateConnectionRequest, TestConnectionRequest } from '@/lib/api/types';
import { connectionApi } from '@/lib/api/connection';
import { toast } from '@/components/ui/use-toast';

interface ConnectionState {
  // 数据
  connections: ConnectionDTO[];
  loading: boolean;
  activeGroupId: string;
  searchKeyword: string;
  // 已打开的连接 ID 集合
  openConnIds: Set<number>;

  // 操作
  fetchConnections: () => Promise<void>;
  setActiveGroup: (groupId: string) => void;
  setSearchKeyword: (keyword: string) => void;
  createConnection: (req: CreateConnectionRequest) => Promise<ConnectionDTO | null>;
  updateConnection: (id: number, req: UpdateConnectionRequest) => Promise<ConnectionDTO | null>;
  deleteConnection: (id: number) => Promise<boolean>;
  testConnection: (req: TestConnectionRequest) => Promise<boolean>;
  openConnection: (id: number) => Promise<boolean>;
  closeConnection: (id: number) => Promise<boolean>;
}

export const useConnectionStore = create<ConnectionState>((set, get) => ({
  connections: [],
  loading: false,
  activeGroupId: '',
  searchKeyword: '',
  openConnIds: new Set(),

  fetchConnections: async () => {
    set({ loading: true });
    try {
      const list = await connectionApi.list({});
      set({ connections: list || [], loading: false });
    } catch (e: any) {
      toast({ title: '获取连接列表失败', description: e.message, variant: 'destructive' });
      set({ loading: false });
    }
  },

  setActiveGroup: (groupId) => set({ activeGroupId: groupId }),
  setSearchKeyword: (keyword) => set({ searchKeyword: keyword }),

  createConnection: async (req) => {
    try {
      const conn = await connectionApi.create(req);
      await get().fetchConnections();
      toast({ title: '创建成功', description: `连接 "${conn.name}" 已创建` });
      return conn;
    } catch (e: any) {
      toast({ title: '创建失败', description: e.message, variant: 'destructive' });
      return null;
    }
  },

  updateConnection: async (id, req) => {
    try {
      const conn = await connectionApi.update(id, req);
      await get().fetchConnections();
      toast({ title: '更新成功', description: `连接 "${conn.name}" 已更新` });
      return conn;
    } catch (e: any) {
      toast({ title: '更新失败', description: e.message, variant: 'destructive' });
      return null;
    }
  },

  deleteConnection: async (id) => {
    try {
      await connectionApi.delete(id);
      // 关闭活跃连接
      const openIds = new Set(get().openConnIds);
      openIds.delete(id);
      set({ openConnIds: openIds });
      await get().fetchConnections();
      toast({ title: '删除成功' });
      return true;
    } catch (e: any) {
      toast({ title: '删除失败', description: e.message, variant: 'destructive' });
      return false;
    }
  },

  testConnection: async (req) => {
    try {
      const result = await connectionApi.test(req);
      if (result.success) {
        toast({ title: '连接成功', description: result.serverVersion || '服务器连接正常' });
      } else {
        toast({ title: '连接失败', description: result.message, variant: 'destructive' });
      }
      return result.success;
    } catch (e: any) {
      toast({ title: '连接失败', description: e.message, variant: 'destructive' });
      return false;
    }
  },

  openConnection: async (id) => {
    try {
      await connectionApi.open(id);
      const openIds = new Set(get().openConnIds);
      openIds.add(id);
      set({ openConnIds: openIds });
      toast({ title: '连接已打开' });
      return true;
    } catch (e: any) {
      toast({ title: '打开连接失败', description: e.message, variant: 'destructive' });
      return false;
    }
  },

  closeConnection: async (id) => {
    try {
      await connectionApi.close(id);
      const openIds = new Set(get().openConnIds);
      openIds.delete(id);
      set({ openConnIds: openIds });
      return true;
    } catch (e: any) {
      toast({ title: '关闭连接失败', description: e.message, variant: 'destructive' });
      return false;
    }
  },
}));
