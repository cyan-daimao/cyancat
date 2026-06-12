import React from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useConnectionStore } from '@/stores/connection';
import { useSchemaStore } from '@/stores/schema';
import {
  ChevronRight, ChevronDown, Database, Server, Table, Eye, Columns,
  FolderOpen, Key, Link2, Loader2, FileType,
} from 'lucide-react';
import type { TreeNode } from '@/stores/schema';
import ObjectTreeContextMenu from './ObjectTreeContextMenu';

const iconMap: Record<string, React.ReactNode> = {
  connection: <Server className="h-4 w-4 text-blue-500" />,
  database: <Database className="h-4 w-4 text-amber-500" />,
  table: <Table className="h-4 w-4 text-green-600" />,
  view: <Eye className="h-4 w-4 text-purple-500" />,
  column: <Columns className="h-4 w-4 text-muted-foreground" />,
  'columns-folder': <FolderOpen className="h-4 w-4 text-blue-400" />,
  'indexes-folder': <FolderOpen className="h-4 w-4 text-orange-400" />,
  'foreign-keys-folder': <FolderOpen className="h-4 w-4 text-pink-400" />,
  index: <Key className="h-4 w-4 text-orange-500" />,
  foreignKey: <Link2 className="h-4 w-4 text-pink-500" />,
};

const ObjectTree: React.FC = () => {
  const { connections, openConnIds, openConnection, closeConnection } = useConnectionStore();
  const {
    trees,
    expandedKeys,
    selectedNode,
    setSelectedNode,
    loadDatabases,
    loadTables,
    loadTableDetail,
    toggleExpand,
    resetTree,
  } = useSchemaStore();
  const [loadingKeys, setLoadingKeys] = React.useState<Set<string>>(new Set());

  const setLoading = (key: string, on: boolean) => {
    setLoadingKeys(prev => {
      const next = new Set(prev);
      if (on) next.add(key); else next.delete(key);
      return next;
    });
  };

  // 点击连接根节点：未打开则 open + 展开 + 拉数据库列表；已打开则切换展开/收起
  const handleConnectionClick = async (node: TreeNode) => {
    const id = node.connID;
    const isOpen = openConnIds.has(id);

    // 未打开 → 先 open
    if (!isOpen) {
      setLoading(node.key, true);
      const ok = await openConnection(id);
      if (!ok) {
        setLoading(node.key, false);
        return;
      }
      // 展开并加载 databases
      if (!expandedKeys.has(node.key)) toggleExpand(node.key);
      await loadDatabases(id);
      setLoading(node.key, false);
      return;
    }

    // 已打开 → 切换展开/收起
    if (expandedKeys.has(node.key)) {
      toggleExpand(node.key);
    } else {
      toggleExpand(node.key);
      // 首次展开但没数据时，按需加载
      if (!trees[id] || trees[id].length === 0) {
        setLoading(node.key, true);
        await loadDatabases(id);
        setLoading(node.key, false);
      }
    }
  };

  // 折叠图标的点击：仅切换 expand，不影响 open/close
  const handleChevronClick = async (e: React.MouseEvent, node: TreeNode) => {
    e.stopPropagation();
    if (node.type === 'connection') {
      // 连接节点的展开走 handleConnectionClick 同样逻辑
      await handleConnectionClick(node);
      return;
    }
    await handleToggle(node);
  };

  // 数据库/表节点：展开 + 懒加载
  const handleToggle = async (node: TreeNode) => {
    const expanding = !expandedKeys.has(node.key);
    toggleExpand(node.key);

    if (!expanding) return; // 收起，不需加载

    if (node.type === 'database' && !node.loaded) {
      setLoading(node.key, true);
      await loadTables(node.connID, node.database!, node.schema || node.database!);
      setLoading(node.key, false);
    } else if (node.type === 'table' && !node.loaded) {
      setLoading(node.key, true);
      await loadTableDetail(node.connID, node.database!, node.schema || node.database!, node.tableName!);
      setLoading(node.key, false);
    }
  };

  const handleNodeClick = async (node: TreeNode, hasChildren: boolean) => {
    setSelectedNode(node);
    if (node.type === 'connection') {
      await handleConnectionClick(node);
    } else if (hasChildren) {
      await handleToggle(node);
    }
  };

  // 双击连接根节点 = 关闭连接（释放资源）
  const handleConnectionDoubleClick = async (e: React.MouseEvent, node: TreeNode) => {
    e.stopPropagation();
    if (openConnIds.has(node.connID)) {
      await closeConnection(node.connID);
      resetTree(node.connID);
    }
  };

  /** 渲染节点附加标签（如字段类型、索引类型、外键引用） */
  const renderNodeBadge = (node: TreeNode): React.ReactNode | null => {
    // 字段节点：从 label 中提取类型（已由 schema store 添加）
    if (node.type === 'column') {
      // label 格式为 "colName (type)"，类型已包含在 label 中
      return null;
    }
    if (node.type === 'index') {
      return (
        <span className="ml-auto text-[10px] px-1 rounded bg-orange-100 dark:bg-orange-900/40 text-orange-600 dark:text-orange-400 leading-tight">
          idx
        </span>
      );
    }
    if (node.type === 'foreignKey') {
      // FK 节点无额外标签，引用信息在 tooltip 中展示
      return null;
    }
    return null;
  };

  const renderNode = (node: TreeNode, depth: number = 0) => {
    const expanded = expandedKeys.has(node.key);
    const isLeaf = node.type === 'column' || node.type === 'view' || node.type === 'index' || node.type === 'foreignKey';
    const hasChildren = !isLeaf;
    const isConn = node.type === 'connection';
    const isLoading = loadingKeys.has(node.key);
    const selected = selectedNode?.key === node.key;

    const nodeElement = (
      <div
        className={`group flex items-center gap-1 px-1 py-0.5 cursor-pointer hover:bg-accent rounded-sm text-sm ${
          selected ? 'bg-accent text-accent-foreground' : ''
        }`}
        style={{ paddingLeft: `${depth * 16 + 4}px` }}
        onClick={() => handleNodeClick(node, hasChildren)}
        onDoubleClick={isConn ? (e) => handleConnectionDoubleClick(e, node) : undefined}
        title={isConn ? '点击展开/收起，双击关闭连接' : undefined}
      >
        {hasChildren ? (
          <span onClick={(e) => handleChevronClick(e, node)} className="shrink-0">
            {isLoading ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : expanded ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronRight className="h-3 w-3" />
            )}
          </span>
        ) : (
          <span className="w-3 shrink-0" />
        )}
        {iconMap[node.type] || <FileType className="h-4 w-4" />}
        <span className="truncate">{node.label}</span>
        {renderNodeBadge(node)}
        {isConn && (
          <span
            className={`ml-auto h-2 w-2 rounded-full ${
              openConnIds.has(node.connID) ? 'bg-green-500' : 'bg-muted-foreground/30'
            }`}
          />
        )}
      </div>
    );

    return (
      <div key={node.key}>
        <ObjectTreeContextMenu node={node}>
          {nodeElement}
        </ObjectTreeContextMenu>
        {expanded && node.children?.map((child) => renderNode(child, depth + 1))}
        {expanded && hasChildren && (!node.children || node.children.length === 0) && !isLoading && node.loaded && (
          <div
            className="text-xs text-muted-foreground italic"
            style={{ paddingLeft: `${(depth + 1) * 16 + 16}px` }}
          >
            (空)
          </div>
        )}
      </div>
    );
  };

  // 构建连接树：每个连接是根节点
  const buildTree = (): TreeNode[] => {
    return connections.map((conn) => {
      const existingChildren = trees[conn.id] || [];
      return {
        key: `conn:${conn.id}`,
        label: conn.name,
        type: 'connection' as const,
        connID: conn.id,
        loaded: existingChildren.length > 0,
        children: existingChildren,
      };
    });
  };

  const treeData = buildTree();

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-3 py-2 border-b border-border">
        <span className="text-xs font-medium text-muted-foreground uppercase">连接</span>
      </div>
      <ScrollArea className="flex-1">
        <div className="py-1">
          {treeData.length === 0 ? (
            <div className="text-xs text-muted-foreground text-center py-4">暂无连接</div>
          ) : (
            treeData.map((node) => renderNode(node))
          )}
        </div>
      </ScrollArea>
    </div>
  );
};

export default ObjectTree;