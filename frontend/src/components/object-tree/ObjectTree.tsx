import React from 'react';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useConnectionStore } from '@/stores/connection';
import { useSchemaStore } from '@/stores/schema';
import {
  ChevronRight, ChevronDown, ChevronLeft, Database, Server, Table, Eye, Columns,
  FolderOpen, Key, Link2, Loader2, FileType, Plus, Search, X,
} from 'lucide-react';
import type { TreeNode } from '@/stores/schema';
import ObjectTreeContextMenu from './ObjectTreeContextMenu';

const iconMap: Record<string, React.ReactNode> = {
  connection: <Server className="h-4 w-4 text-blue-500" />,
  database: <Database className="h-4 w-4 text-amber-500" />,
  schema: <FolderOpen className="h-4 w-4 text-cyan-500" />,
  table: <Table className="h-4 w-4 text-green-600" />,
  view: <Eye className="h-4 w-4 text-purple-500" />,
  column: <Columns className="h-4 w-4 text-muted-foreground" />,
  'columns-folder': <FolderOpen className="h-4 w-4 text-blue-400" />,
  'indexes-folder': <FolderOpen className="h-4 w-4 text-orange-400" />,
  'foreign-keys-folder': <FolderOpen className="h-4 w-4 text-pink-400" />,
  index: <Key className="h-4 w-4 text-orange-500" />,
  foreignKey: <Link2 className="h-4 w-4 text-pink-500" />,
};

interface ObjectTreeProps {
  onCreateConnection?: () => void;
  onCollapse?: () => void;
}

const ObjectTree: React.FC<ObjectTreeProps> = ({ onCreateConnection, onCollapse }) => {
  const { connections, openConnIds, openConnection, closeConnection } = useConnectionStore();
  const {
    trees,
    expandedKeys,
    selectedNode,
    setSelectedNode,
    loadDatabases,
    loadSchemas,
    loadTables,
    loadTableDetail,
    toggleExpand,
    resetTree,
  } = useSchemaStore();
  const [loadingKeys, setLoadingKeys] = React.useState<Set<string>>(new Set());

  // 搜索关键字（前端过滤数据库 + 表/视图名称）
  const [searchKeyword, setSearchKeyword] = React.useState('');
  const trimmedKeyword = searchKeyword.trim().toLowerCase();
  const isSearching = trimmedKeyword.length > 0;

  const setLoading = (key: string, on: boolean) => {
    setLoadingKeys(prev => {
      const next = new Set(prev);
      if (on) next.add(key); else next.delete(key);
      return next;
    });
  };

  const getConnectionType = (connID: number) =>
    connections.find(conn => conn.id === connID)?.type;

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
      if (getConnectionType(node.connID) === 'postgres') {
        await loadSchemas(node.connID, node.database!);
      } else {
        await loadTables(node.connID, node.database!, node.schema || node.database!);
      }
      setLoading(node.key, false);
    } else if (node.type === 'schema' && !node.loaded) {
      setLoading(node.key, true);
      await loadTables(node.connID, node.database!, node.schema || 'public');
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

  /** 在 label 中高亮匹配片段 */
  const renderLabel = (label: string, matchable: boolean): React.ReactNode => {
    if (!isSearching || !matchable || !trimmedKeyword) return label;
    const lower = label.toLowerCase();
    const idx = lower.indexOf(trimmedKeyword);
    if (idx < 0) return label;
    const end = idx + trimmedKeyword.length;
    return (
      <>
        {label.slice(0, idx)}
        <mark className="bg-yellow-200 dark:bg-yellow-500/40 text-foreground rounded-sm px-0.5">
          {label.slice(idx, end)}
        </mark>
        {label.slice(end)}
      </>
    );
  };

  const renderNode = (node: TreeNode, depth: number = 0) => {
    const expanded = isExpanded(node.key);
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
        <span className="truncate">{renderLabel(node.label, isMatchableType(node.type))}</span>
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

  /**
   * 节点自身是否匹配搜索关键字。
   * 仅匹配数据库 / schema / 表 / 视图（用户需求：数据库 + 表）。
   * 连接节点也参与匹配（按连接名）。
   */
  const isMatchableType = (t: TreeNode['type']) =>
    t === 'connection' || t === 'database' || t === 'schema' || t === 'table' || t === 'view';

  const nodeMatches = (node: TreeNode): boolean => {
    if (!isMatchableType(node.type)) return false;
    // 表节点：tableName 更精准（label 含原始名即可）
    const hay = (node.tableName || node.label || '').toLowerCase();
    return hay.includes(trimmedKeyword);
  };

  /**
   * 过滤树：保留自身匹配或拥有匹配后代的节点。
   * 同时收集需要自动展开的祖先 key，用于搜索时撑开层级。
   */
  const filterTree = (
    nodes: TreeNode[],
    autoExpand: Set<string>,
  ): TreeNode[] => {
    const out: TreeNode[] = [];
    for (const node of nodes) {
      const selfMatch = nodeMatches(node);
      const filteredChildren = node.children
        ? filterTree(node.children, autoExpand)
        : undefined;
      const hasChildMatch = !!filteredChildren && filteredChildren.length > 0;

      if (selfMatch || hasChildMatch) {
        if (hasChildMatch) autoExpand.add(node.key);
        out.push({
          ...node,
          // 自身匹配时仍展示其原始子节点（不剪枝），方便用户继续浏览
          children: selfMatch ? node.children : filteredChildren,
        });
      }
    }
    return out;
  };

  const rawTree = buildTree();
  const autoExpandKeys = React.useMemo(() => new Set<string>(), [trimmedKeyword]);
  const treeData = isSearching ? filterTree(rawTree, autoExpandKeys) : rawTree;

  // 节点是否应被视为展开（搜索态下，匹配子树的祖先一律视为展开）
  const isExpanded = (key: string) =>
    expandedKeys.has(key) || (isSearching && autoExpandKeys.has(key));

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-3 py-2 border-b border-border">
        <span className="text-xs font-medium text-muted-foreground uppercase">连接</span>
        <div className="flex items-center gap-0.5">
          <Button
            variant="ghost"
            size="icon"
            className="h-5 w-5"
            onClick={onCreateConnection}
            title="新增数据源"
          >
            <Plus className="h-3 w-3" />
          </Button>
          {onCollapse && (
            <Button
              variant="ghost"
              size="icon"
              className="h-5 w-5"
              onClick={onCollapse}
              title="收起侧边栏"
            >
              <ChevronLeft className="h-3 w-3" />
            </Button>
          )}
        </div>
      </div>

      {/* 搜索栏：前端过滤数据库 / 表 / 视图 */}
      <div className="px-2 py-1.5 border-b border-border shrink-0">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3 w-3 text-muted-foreground pointer-events-none" />
          <Input
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
            placeholder="搜索数据库 / 表"
            className="h-7 pl-7 pr-7 text-xs"
          />
          {searchKeyword && (
            <button
              type="button"
              onClick={() => setSearchKeyword('')}
              className="absolute right-1.5 top-1/2 -translate-y-1/2 inline-flex h-4 w-4 items-center justify-center rounded text-muted-foreground hover:bg-accent hover:text-foreground"
              title="清除"
            >
              <X className="h-3 w-3" />
            </button>
          )}
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="py-1">
          {treeData.length === 0 ? (
            <div className="text-xs text-muted-foreground text-center py-4">
              {isSearching ? '无匹配结果' : '暂无连接'}
            </div>
          ) : (
            treeData.map((node) => renderNode(node))
          )}
        </div>
      </ScrollArea>
    </div>
  );
};

export default ObjectTree;
