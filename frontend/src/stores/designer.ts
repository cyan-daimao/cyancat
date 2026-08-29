import { create } from 'zustand';
import type { TableDetailDTO } from '@/lib/api/types';

// 设计器上下文：打开 Table Designer 时传入
export interface TableDesignerContext {
  mode: 'create' | 'edit' | 'view';
  connID: number;
  database: string;
  schema: string;
  tableName?: string;       // edit/view 模式必填
  focusField?: string;      // 字段聚焦
  focusTab?: 'fields' | 'indexes' | 'foreignKeys' | 'options';
  initialDetail?: TableDetailDTO | null;
}

// 创建数据库对话框上下文
export interface CreateDatabaseContext {
  connID: number;
}

// DDL 查看器上下文
export interface DDLViewerContext {
  connID: number;
  database: string;
  schema: string;
  table: string;
}

interface DesignerState {
  // 新建数据库对话框
  createDatabaseOpen: boolean;
  createDatabaseContext: CreateDatabaseContext | null;
  openCreateDatabase: (ctx: CreateDatabaseContext) => void;
  closeCreateDatabase: () => void;

  // 表设计器对话框
  tableDesignerOpen: boolean;
  tableDesignerContext: TableDesignerContext | null;
  openTableDesigner: (ctx: TableDesignerContext) => void;
  closeTableDesigner: () => void;

  // 删除表确认弹窗
  dropTableConfirmOpen: boolean;
  dropTableContext: { connID: number; database: string; schema: string; tableName: string } | null;
  openDropTableConfirm: (ctx: { connID: number; database: string; schema: string; tableName: string }) => void;
  closeDropTableConfirm: () => void;

  // 删除数据库确认弹窗
  dropDatabaseConfirmOpen: boolean;
  dropDatabaseContext: { connID: number; name: string } | null;
  openDropDatabaseConfirm: (ctx: { connID: number; name: string }) => void;
  closeDropDatabaseConfirm: () => void;

  // 删除连接确认弹窗
  deleteConnectionConfirmOpen: boolean;
  deleteConnectionContext: { connID: number; name: string } | null;
  openDeleteConnectionConfirm: (ctx: { connID: number; name: string }) => void;
  closeDeleteConnectionConfirm: () => void;

  // DDL 查看器对话框
  ddlViewerOpen: boolean;
  ddlViewerContext: DDLViewerContext | null;
  openDDLViewer: (ctx: DDLViewerContext) => void;
  closeDDLViewer: () => void;

  // MCP Server 对话框（全局单例，无需 connID）
  mcpServerDialogOpen: boolean;
  openMcpServerDialog: () => void;
  closeMcpServerDialog: () => void;
}

export const useDesignerStore = create<DesignerState>((set) => ({
  createDatabaseOpen: false,
  createDatabaseContext: null,
  openCreateDatabase: (ctx) => set({ createDatabaseOpen: true, createDatabaseContext: ctx }),
  closeCreateDatabase: () => set({ createDatabaseOpen: false, createDatabaseContext: null }),

  tableDesignerOpen: false,
  tableDesignerContext: null,
  openTableDesigner: (ctx) => set({ tableDesignerOpen: true, tableDesignerContext: ctx }),
  closeTableDesigner: () => set({ tableDesignerOpen: false, tableDesignerContext: null }),

  ddlViewerOpen: false,
  ddlViewerContext: null,
  openDDLViewer: (ctx) => set({ ddlViewerOpen: true, ddlViewerContext: ctx }),
  closeDDLViewer: () => set({ ddlViewerOpen: false, ddlViewerContext: null }),

  mcpServerDialogOpen: false,
  openMcpServerDialog: () => set({ mcpServerDialogOpen: true }),
  closeMcpServerDialog: () => set({ mcpServerDialogOpen: false }),

  dropTableConfirmOpen: false,
  dropTableContext: null,
  openDropTableConfirm: (ctx) => set({ dropTableConfirmOpen: true, dropTableContext: ctx }),
  closeDropTableConfirm: () => set({ dropTableConfirmOpen: false, dropTableContext: null }),

  dropDatabaseConfirmOpen: false,
  dropDatabaseContext: null,
  openDropDatabaseConfirm: (ctx) => set({ dropDatabaseConfirmOpen: true, dropDatabaseContext: ctx }),
  closeDropDatabaseConfirm: () => set({ dropDatabaseConfirmOpen: false, dropDatabaseContext: null }),

  deleteConnectionConfirmOpen: false,
  deleteConnectionContext: null,
  openDeleteConnectionConfirm: (ctx) => set({ deleteConnectionConfirmOpen: true, deleteConnectionContext: ctx }),
  closeDeleteConnectionConfirm: () => set({ deleteConnectionConfirmOpen: false, deleteConnectionContext: null }),
}));
