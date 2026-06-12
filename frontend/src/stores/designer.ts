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
  focusTab?: 'fields' | 'indexes' | 'foreignKeys' | 'options' | 'ddl';
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

  // DDL 查看器对话框
  ddlViewerOpen: boolean;
  ddlViewerContext: DDLViewerContext | null;
  openDDLViewer: (ctx: DDLViewerContext) => void;
  closeDDLViewer: () => void;
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

  dropTableConfirmOpen: false,
  dropTableContext: null,
  openDropTableConfirm: (ctx) => set({ dropTableConfirmOpen: true, dropTableContext: ctx }),
  closeDropTableConfirm: () => set({ dropTableConfirmOpen: false, dropTableContext: null }),
}));
