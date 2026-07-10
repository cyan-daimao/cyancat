import React from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useConnectionStore } from '@/stores/connection';
import { useSchemaStore, DEFAULT_TABLE_PAGE_SIZE, type TreeNode } from '@/stores/schema';
import {
  ChevronRight, ChevronDown, ChevronLeft, Database, Server, Table, Eye, Columns,
  FolderOpen, Key, Link2, Loader2, FileType, Plus, Search, X,
} from 'lucide-react';
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

import type { ConnectionDTO } from '@/lib/api/types';

const ROW_HEIGHT = 24;
const OVERSCAN = 10;

interface ObjectTreeProps {
  onCreateConnection?: () => void;
  onShowProperties?: (conn: ConnectionDTO) => void;
  onCollapse?: () => void;
}

const ObjectTree: React.FC<ObjectTreeProps> = ({ onCreateConnection, onShowProperties, onCollapse }) => {
  const { connections, openConnIds, openConnection, closeConnection } = useConnectionStore();
  const {
    trees,
    expandedKeys,
    selectedNode,
    searchMap,
    setSelectedNode,
    loadDatabases,
    loadSchemas,
    loadTables,
    loadMoreTables,
    loadMoreViews,
    searchTables,
    loadTableDetail,
    toggleExpand,
    resetTree,
  } = useSchemaStore();
  const [loadingKeys, setLoadingKeys] = React.useState<Set<string>>(new Set());

  // 搜索关键字：输入框立即响应，但搜索逻辑使用 debounce，避免每个字符都触发整树重算
  const [searchKeyword, setSearchKeyword] = React.useState('');
  const trimmedKeyword = searchKeyword.trim().toLowerCase();
  const [debouncedKeyword, setDebouncedKeyword] = React.useState('');
  const isSearching = debouncedKeyword.length > 0;
  const searchTimeoutRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(() => {
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
      searchTimeoutRef.current = null;
    }
    searchTimeoutRef.current = setTimeout(() => {
      setDebouncedKeyword(trimmedKeyword);
    }, 300);
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
        searchTimeoutRef.current = null;
      }
    };
  }, [trimmedKeyword]);

  const setLoading = (key: string, on: boolean) => {
    setLoadingKeys(prev => {
      const next = new Set(prev);
      if (on) next.add(key); else next.delete(key);
      return next;
    });
  };

  const getConnectionType = (connID: number) =>
    connections.find(conn => conn.id === connID)?.type;

  const needsSchemaLayer = (connID: number) => {
    const t = getConnectionType(connID);
    return t === 'postgres' || t === 'starrocks';
  };

  const handleConnectionClick = async (node: TreeNode) => {
    const id = node.connID;
    const isOpen = openConnIds.has(id);

    if (!isOpen) {
      setLoading(node.key, true);
      const ok = await openConnection(id);
      if (!ok) {
        setLoading(node.key, false);
        return;
      }
      if (!expandedKeys.has(node.key)) toggleExpand(node.key);
      await loadDatabases(id);
      setLoading(node.key, false);
      return;
    }

    if (expandedKeys.has(node.key)) {
      toggleExpand(node.key);
    } else {
      toggleExpand(node.key);
      if (!trees[id] || trees[id].length === 0) {
        setLoading(node.key, true);
        await loadDatabases(id);
        setLoading(node.key, false);
      }
    }
  };

  const handleChevronClick = async (e: React.MouseEvent, node: TreeNode) => {
    e.stopPropagation();
    if (node.type === 'connection') {
      await handleConnectionClick(node);
      return;
    }
    await handleToggle(node);
  };

  const handleToggle = async (node: TreeNode) => {
    const expanding = !expandedKeys.has(node.key);
    toggleExpand(node.key);

    if (!expanding) return;

    if (node.type === 'database' && !node.loaded) {
      setLoading(node.key, true);
      if (needsSchemaLayer(node.connID)) {
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

  const handleConnectionDoubleClick = async (e: React.MouseEvent, node: TreeNode) => {
    e.stopPropagation();
    if (openConnIds.has(node.connID)) {
      await closeConnection(node.connID);
      resetTree(node.connID);
    }
  };

  const renderNodeBadge = (node: TreeNode): React.ReactNode | null => {
    if (node.type === 'column') return null;
    if (node.type === 'index') {
      return (
        <span className="ml-auto text-[10px] px-1 rounded bg-orange-100 dark:bg-orange-900/40 text-orange-600 dark:text-orange-400 leading-tight">
          idx
        </span>
      );
    }
    if (node.type === 'foreignKey') return null;
    return null;
  };

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

  const rawTree = React.useMemo(() => {
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
  }, [connections, trees]);

  // 扁平化可见节点（搜索时强制展开 connection 以显示命中结果）
  const flatItems = React.useMemo(() => {
    const flattenVisible = (nodes: TreeNode[], depth = 0): { node: TreeNode; depth: number }[] => {
      const out: { node: TreeNode; depth: number }[] = [];
      for (const node of nodes) {
        out.push({ node, depth });
        const expanded = expandedKeys.has(node.key) || (isSearching && node.type === 'connection');
        if (expanded && node.children) {
          out.push(...flattenVisible(node.children, depth + 1));
        }
      }
      return out;
    };
    return flattenVisible(rawTree);
  }, [rawTree, expandedKeys, isSearching]);

  // 搜索触发：用 ref 保存最新状态，避免 useEffect 依赖数组导致闭包过时
  const connectionsRef = React.useRef(connections);
  const rawTreeRef = React.useRef(rawTree);
  const expandedKeysRef = React.useRef(expandedKeys);
  const searchTablesRef = React.useRef(searchTables);
  React.useEffect(() => {
    connectionsRef.current = connections;
    rawTreeRef.current = rawTree;
    expandedKeysRef.current = expandedKeys;
    searchTablesRef.current = searchTables;
  });
  React.useEffect(() => {
    if (!isSearching) return;
    const runSearch = async () => {
      const conns = connectionsRef.current;
      const tree = rawTreeRef.current;
      const expanded = expandedKeysRef.current;
      const doSearch = searchTablesRef.current;
      for (const conn of conns) {
        const connNode = tree.find(n => n.connID === conn.id);
        if (!connNode || !connNode.children) continue;
        for (const dbNode of connNode.children) {
          if (!expanded.has(dbNode.key) || !dbNode.database) continue;
          if (needsSchemaLayer(conn.id)) {
            for (const schemaNode of dbNode.children || []) {
              if (expanded.has(schemaNode.key) && schemaNode.schema) {
                await doSearch(conn.id, dbNode.database!, schemaNode.schema!, debouncedKeyword);
              }
            }
          } else {
            await doSearch(conn.id, dbNode.database!, dbNode.database!, debouncedKeyword);
          }
        }
      }
    };
    runSearch();
  }, [debouncedKeyword, isSearching]);

  // 搜索结果覆盖：把匹配节点替换进 flatItems
  const visibleItems = React.useMemo(() => {
    if (!isSearching) return flatItems;

    const out: { node: TreeNode; depth: number }[] = [];
    for (const { node, depth } of flatItems) {
      out.push({ node, depth });
      if ((node.type === 'database' || node.type === 'schema') && expandedKeys.has(node.key)) {
        const searchState = searchMap[node.key];
        if (searchState?.results.length) {
          for (const child of searchState.results) {
            out.push({ node: child, depth: depth + 1 });
          }
        }
      }
    }
    return out;
  }, [flatItems, isSearching, expandedKeys, searchMap]);

  // 虚拟滚动
  const parentRef = React.useRef<HTMLDivElement>(null);
  const virtualizer = useVirtualizer({
    count: visibleItems.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  });
  const virtualItems = virtualizer.getVirtualItems();

  // 滚动到底加载更多
  const lastScrollTop = React.useRef(0);
  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const target = e.currentTarget;
    const { scrollTop, scrollHeight, clientHeight } = target;
    const isScrollingDown = scrollTop > lastScrollTop.current;
    lastScrollTop.current = scrollTop;

    if (!isScrollingDown) return;
    if (scrollHeight - scrollTop - clientHeight > ROW_HEIGHT * 2) return;

    // 找到当前可视区末尾附近的 database/schema 节点，尝试加载更多
    for (let i = virtualItems.length - 1; i >= 0; i--) {
      const item = virtualItems[i];
      const { node } = visibleItems[item.index];
      if (node.type === 'database' || node.type === 'schema') {
        if (node.hasMoreTables) {
          loadMoreTables(node.key);
          break;
        }
        if (node.hasMoreViews) {
          loadMoreViews(node.key);
          break;
        }
      }
    }
  };

  const isLeaf = (node: TreeNode) =>
    node.type === 'column' || node.type === 'view' || node.type === 'index' || node.type === 'foreignKey';

  const isMatchableType = (t: TreeNode['type']) =>
    t === 'connection' || t === 'database' || t === 'schema' || t === 'table' || t === 'view';

  const renderRow = (index: number, style: React.CSSProperties) => {
    const { node, depth } = visibleItems[index];
    const expanded = expandedKeys.has(node.key);
    const hasChildren = !isLeaf(node);
    const isLoading = loadingKeys.has(node.key);
    const selected = selectedNode?.key === node.key;

    return (
      <ObjectTreeContextMenu key={node.key} node={node} onShowProperties={onShowProperties}>
        <div
          className={`group flex items-center gap-1 px-1 cursor-pointer hover:bg-accent rounded-sm text-sm ${
            selected ? 'bg-accent text-accent-foreground' : ''
          }`}
          style={{ ...style, paddingLeft: `${depth * 16 + 4}px` }}
          onClick={() => handleNodeClick(node, hasChildren)}
          onDoubleClick={node.type === 'connection' ? (e) => handleConnectionDoubleClick(e, node) : undefined}
          title={node.type === 'connection' ? '点击展开/收起，双击关闭连接' : undefined}
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
          {node.type === 'connection' && (
            <span
              className={`ml-auto h-2 w-2 rounded-full ${
                openConnIds.has(node.connID) ? 'bg-green-500' : 'bg-muted-foreground/30'
              }`}
            />
          )}
        </div>
      </ObjectTreeContextMenu>
    );
  };

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

      {/* 搜索栏 */}
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

      <div ref={parentRef} className="flex-1 overflow-auto" onScroll={handleScroll}>
        {visibleItems.length === 0 ? (
          <div className="text-xs text-muted-foreground text-center py-4">
            {isSearching ? '无匹配结果' : '暂无连接'}
          </div>
        ) : (
          <div style={{ height: `${virtualizer.getTotalSize()}px`, position: 'relative' }}>
            {virtualItems.map((virtualItem) => (
              <div
                key={virtualItem.key}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualItem.start}px)`,
                }}
              >
                {renderRow(virtualItem.index, { height: `${virtualItem.size}px` })}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ObjectTree;
