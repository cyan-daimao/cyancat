import React from 'react';
import {
  ContextMenu,
  ContextMenuTrigger,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuSub,
  ContextMenuSubTrigger,
  ContextMenuSubContent,
} from '@/components/ui/context-menu';
import * as Icons from 'lucide-react';
import { useSchemaStore } from '@/stores/schema';
import { useConnectionStore } from '@/stores/connection';
import { useDesignerStore } from '@/stores/designer';
import { useQueryStore } from '@/stores/query';
import { schemaApi } from '@/lib/api/schema';
import { toast } from '@/components/ui/use-toast';
import { cn } from '@/lib/utils';
import type { TreeNode } from '@/stores/schema';
import { getMenuItemsForNode, type MenuItemDef, type MenuContext } from './context-menus';

interface ObjectTreeContextMenuProps {
  node: TreeNode;
  children: React.ReactNode;
}

/** 渲染图标：从 lucide-react 动态查找 */
function renderIcon(name?: string): React.ReactNode {
  if (!name) return null;
  const Comp = (Icons as Record<string, unknown>)[name] as
    | React.ComponentType<{ className?: string }>
    | undefined;
  if (!Comp) return null;
  return <Comp className="h-4 w-4" />;
}

const ObjectTreeContextMenu: React.FC<ObjectTreeContextMenuProps> = ({ node, children }) => {
  const { setSelectedNode, loadDatabases, loadTables, loadTableDetail, resetTree } = useSchemaStore();
  const { openConnIds, openConnection, closeConnection } = useConnectionStore();
  const { openCreateDatabase, openTableDesigner, openDDLViewer, openDropTableConfirm } = useDesignerStore();
  const { addEmptyResult, execute } = useQueryStore();

  const isConnOpen = openConnIds.has(node.connID);
  const menuItems = getMenuItemsForNode(node);
  const menuCtx: MenuContext = { node, isConnOpen };

  // ---- 通用工具 ----
  const copyText = async (text: string, title = '已复制') => {
    try {
      await navigator.clipboard.writeText(text);
      toast({ title, description: text });
    } catch (e: any) {
      toast({ title: '复制失败', description: e.message, variant: 'destructive' });
    }
  };

  const qualifiedName = (): string => {
    const parts: string[] = [];
    if (node.database) parts.push(node.database);
    if (node.schema && node.schema !== node.database) parts.push(node.schema);
    if (node.tableName) parts.push(node.tableName);
    if (node.columnName) parts.push(node.columnName);
    if (parts.length === 0) parts.push(node.label);
    return parts.join('.');
  };

  const quotedName = (name: string): string => `\`${name}\``;

  /** 统一 action 派发 */
  const dispatchAction = async (action: string) => {
    setSelectedNode(node);
    switch (action) {
      // ---- 连接 ----
      case 'open-connection': {
        if (!isConnOpen) {
          const ok = await openConnection(node.connID);
          if (ok) await loadDatabases(node.connID);
        }
        return;
      }
      case 'close-connection': {
        if (isConnOpen) {
          await closeConnection(node.connID);
          resetTree(node.connID);
        }
        return;
      }
      case 'new-database': {
        openCreateDatabase({ connID: node.connID });
        return;
      }
      case 'connection-properties': {
        toast({ title: '连接属性', description: '即将在后续版本支持' });
        return;
      }

      // ---- 数据库 ----
      case 'set-current-database': {
        toast({ title: '已切换数据库', description: node.database });
        return;
      }
      case 'database-properties': {
        toast({ title: '数据库属性', description: '即将在后续版本支持' });
        return;
      }

      // ---- 表 ----
      case 'new-table': {
        openTableDesigner({
          mode: 'create',
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
        });
        return;
      }
      case 'design-table': {
        openTableDesigner({
          mode: 'edit',
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          tableName: node.tableName,
        });
        return;
      }
      case 'open-table': {
        const sql = `SELECT * FROM ${quotedName(node.database!)}.${quotedName(node.tableName!)} LIMIT 500;`;
        await execute({
          connID: node.connID,
          sql,
          maxRows: 500,
          database: node.database,
          schema: node.schema,
        });
        return;
      }
      case 'view-ddl': {
        openDDLViewer({
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          table: node.tableName!,
        });
        return;
      }
      case 'copy-ddl': {
        try {
          const ddl = await schemaApi.getCreateTableDDL(
            node.connID,
            node.database!,
            node.schema || node.database!,
            node.tableName!,
          );
          await copyText(ddl, 'DDL 已复制');
        } catch (e: any) {
          toast({ title: '获取 DDL 失败', description: e.message, variant: 'destructive' });
        }
        return;
      }
      case 'generate-sql-select': {
        const sql = `SELECT * FROM ${quotedName(node.database!)}.${quotedName(node.tableName!)} LIMIT 100;`;
        await copyText(sql, 'SELECT SQL 已复制');
        return;
      }
      case 'drop-table': {
        openDropTableConfirm({
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          tableName: node.tableName!,
        });
        return;
      }

      // ---- 视图 ----
      case 'open-view': {
        const sql = `SELECT * FROM ${quotedName(node.database!)}.${quotedName(node.tableName!)} LIMIT 500;`;
        await execute({
          connID: node.connID,
          sql,
          maxRows: 500,
          database: node.database,
          schema: node.schema,
        });
        return;
      }

      // ---- 字段 ----
      case 'edit-column':
      case 'add-column':
      case 'insert-column':
      case 'rename-column': {
        openTableDesigner({
          mode: 'edit',
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          tableName: node.tableName!,
          focusTab: 'fields',
          focusField: node.columnName,
        });
        return;
      }
      case 'drop-column': {
        openTableDesigner({
          mode: 'edit',
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          tableName: node.tableName!,
          focusTab: 'fields',
          focusField: node.columnName,
        });
        toast({ title: '在表设计器中确认删除字段' });
        return;
      }
      case 'set-primary-key':
      case 'drop-primary-key': {
        openTableDesigner({
          mode: 'edit',
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          tableName: node.tableName!,
          focusTab: 'fields',
          focusField: node.columnName,
        });
        return;
      }
      case 'copy-quoted-name': {
        await copyText(quotedName(node.columnName || node.label));
        return;
      }

      // ---- 索引 / 外键 ----
      case 'new-index':
      case 'edit-index':
      case 'drop-index': {
        openTableDesigner({
          mode: 'edit',
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          tableName: node.tableName!,
          focusTab: 'indexes',
        });
        return;
      }
      case 'new-foreign-key':
      case 'edit-foreign-key':
      case 'drop-foreign-key': {
        openTableDesigner({
          mode: 'edit',
          connID: node.connID,
          database: node.database!,
          schema: node.schema || node.database!,
          tableName: node.tableName!,
          focusTab: 'foreignKeys',
        });
        return;
      }

      // ---- 通用 ----
      case 'new-query': {
        addEmptyResult();
        toast({ title: '已新建查询', description: '请在 SQL 编辑器中输入' });
        return;
      }
      case 'refresh': {
        if (node.type === 'connection') {
          await loadDatabases(node.connID);
        } else if (node.type === 'database') {
          await loadTables(node.connID, node.database!, node.schema || node.database!);
        } else if (node.type === 'table') {
          await loadTableDetail(
            node.connID,
            node.database!,
            node.schema || node.database!,
            node.tableName!,
          );
        }
        return;
      }
      case 'refresh-table': {
        await loadTableDetail(
          node.connID,
          node.database!,
          node.schema || node.database!,
          node.tableName!,
        );
        return;
      }
      case 'copy-name': {
        await copyText(node.columnName || node.tableName || node.database || node.label);
        return;
      }
      case 'copy-full-name': {
        await copyText(qualifiedName());
        return;
      }

      default: {
        toast({ title: '敬请期待', description: action });
      }
    }
  };

  /** 渲染一个菜单项（叶子或子菜单） */
  const renderItem = (item: MenuItemDef, idx: number): React.ReactNode => {
    if (item.separator) {
      return <ContextMenuSeparator key={`sep-${idx}`} />;
    }

    const disabled = item.disabled ? item.disabled(menuCtx) : false;
    const icon = renderIcon(item.icon);

    // 子菜单
    if (item.children && item.children.length > 0) {
      return (
        <ContextMenuSub key={`sub-${idx}-${item.label}`}>
          <ContextMenuSubTrigger
            className={cn(item.danger && 'text-red-500 focus:text-red-500')}
            disabled={disabled}
          >
            {icon}
            <span>{item.label}</span>
          </ContextMenuSubTrigger>
          <ContextMenuSubContent className="min-w-[160px]">
            {item.children.map((c, ci) => renderItem(c, ci))}
          </ContextMenuSubContent>
        </ContextMenuSub>
      );
    }

    return (
      <ContextMenuItem
        key={`item-${idx}-${item.action}`}
        disabled={disabled}
        onClick={() => item.action && dispatchAction(item.action)}
        className={cn(item.danger && 'text-red-500 focus:text-red-500')}
      >
        {icon}
        <span>{item.label}</span>
      </ContextMenuItem>
    );
  };

  // 没有菜单项时，直接返回 children（不包 trigger）
  if (menuItems.length === 0) {
    return <>{children}</>;
  }

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild onContextMenu={() => setSelectedNode(node)}>
        {children}
      </ContextMenuTrigger>
      <ContextMenuContent className="min-w-[200px]">
        {menuItems.map((item, idx) => renderItem(item, idx))}
      </ContextMenuContent>
    </ContextMenu>
  );
};

export default ObjectTreeContextMenu;
