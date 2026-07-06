import React from 'react';
import SqlEditor from './SqlEditor';
import ResultPanel from '@/components/data-table/ResultPanel';
import { useQueryStore, type QueryTabContext } from '@/stores/query';
import { useSchemaStore } from '@/stores/schema';
import { useConnectionStore } from '@/stores/connection';
import { Plus, X } from 'lucide-react';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { Tooltip, TooltipTrigger, TooltipContent } from '@/components/ui/tooltip';
import { toast } from '@/components/ui/use-toast';

const copySql = async (sql?: string) => {
  if (!sql) return;
  try {
    await navigator.clipboard.writeText(sql);
    toast({ title: '已复制 SQL' });
  } catch (e: any) {
    toast({ title: '复制失败', description: e?.message, variant: 'destructive' });
  }
};

const MIN_EDITOR_HEIGHT = 80;
const MAX_EDITOR_HEIGHT = 600;
const DEFAULT_EDITOR_HEIGHT = 200;

const SqlWorkspace: React.FC = () => {
  const {
    queryTabs,
    activeQueryTabId,
    createQueryTab,
    closeQueryTab,
    closeOtherQueryTabs,
    closeQueryTabsToRight,
    closeQueryTabsToLeft,
    setActiveQueryTab,
    updateActiveTabSql,
    updateActiveTabContext,
    setActiveResult,
    closeResult,
    closeOtherResults,
    closeAllResults,
  } = useQueryStore();
  const selectedNode = useSchemaStore(s => s.selectedNode);
  const connections = useConnectionStore(s => s.connections);
  const activeTab = queryTabs.find(tab => tab.id === activeQueryTabId) || queryTabs[0];
  const results = activeTab?.results || [];
  const activeResultIndex = activeTab?.activeResultIndex || 0;

  const selectedContext = React.useMemo<QueryTabContext | null>(() => {
    if (!selectedNode) return null;
    const conn = connections.find(c => c.id === selectedNode.connID);
    const database = conn?.type === 'sqlite' ? (selectedNode.database || 'main') : selectedNode.database;
    const schema = conn?.type === 'sqlite' ? (selectedNode.schema || 'main') : selectedNode.schema;
    const parts = [conn?.name || `连接 ${selectedNode.connID}`];
    if (database) parts.push(database);
    if (schema) parts.push(schema);
    return {
      connID: selectedNode.connID,
      connectionType: conn?.type,
      database,
      schema,
      contextLabel: parts.join(' / '),
    };
  }, [selectedNode, connections]);

  React.useEffect(() => {
    if (selectedContext) {
      updateActiveTabContext(selectedContext);
    }
  }, [selectedContext, updateActiveTabContext]);

  // 拖拽分割线状态
  const [editorHeight, setEditorHeight] = React.useState(DEFAULT_EDITOR_HEIGHT);
  const dragging = React.useRef(false);
  const startY = React.useRef(0);
  const startHeight = React.useRef(0);

  const onMouseDown = React.useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    dragging.current = true;
    startY.current = e.clientY;
    startHeight.current = editorHeight;
    document.body.style.cursor = 'row-resize';
    document.body.style.userSelect = 'none';

    const onMouseMove = (ev: MouseEvent) => {
      if (!dragging.current) return;
      const delta = ev.clientY - startY.current;
      const newHeight = Math.min(MAX_EDITOR_HEIGHT, Math.max(MIN_EDITOR_HEIGHT, startHeight.current + delta));
      setEditorHeight(newHeight);
    };

    const onMouseUp = () => {
      dragging.current = false;
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
    };

    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  }, [editorHeight]);

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* SQL 标签页 */}
      <div className="flex items-center border-b border-border shrink-0 bg-muted/30 overflow-x-auto">
        {queryTabs.map((tab, index) => (
          <ContextMenu key={tab.id}>
            <ContextMenuTrigger asChild onContextMenu={() => setActiveQueryTab(tab.id)}>
              <div
                className={
                  'flex items-center gap-1 px-3 py-1.5 text-xs border-r border-border cursor-pointer whitespace-nowrap group ' +
                  (tab.id === activeQueryTabId
                    ? 'bg-background text-foreground border-b-2 border-b-primary'
                    : 'text-muted-foreground hover:bg-accent/30 hover:text-foreground')
                }
                onClick={() => setActiveQueryTab(tab.id)}
              >
                <span>{tab.title}</span>
                {tab.results.length > 0 && (
                  <span className="text-muted-foreground">({tab.results.length})</span>
                )}
                <button
                  className="ml-1 inline-flex h-4 w-4 items-center justify-center rounded text-foreground/70 hover:bg-accent hover:text-foreground"
                  onClick={(e) => { e.stopPropagation(); closeQueryTab(tab.id); }}
                  title="关闭标签页"
                >
                  <X className="h-3 w-3" />
                </button>
              </div>
            </ContextMenuTrigger>
            <ContextMenuContent>
              <ContextMenuItem onClick={() => closeQueryTab(tab.id)}>
                关闭
              </ContextMenuItem>
              <ContextMenuItem
                onClick={() => closeOtherQueryTabs(tab.id)}
                disabled={queryTabs.length <= 1}
              >
                关闭其他
              </ContextMenuItem>
              <ContextMenuSeparator />
              <ContextMenuItem
                onClick={() => closeQueryTabsToRight(tab.id)}
                disabled={index === queryTabs.length - 1}
              >
                关闭右侧
              </ContextMenuItem>
              <ContextMenuItem
                onClick={() => closeQueryTabsToLeft(tab.id)}
                disabled={index === 0}
              >
                关闭左侧
              </ContextMenuItem>
            </ContextMenuContent>
          </ContextMenu>
        ))}
        <button
          className="h-7 w-7 inline-flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-accent/40 shrink-0"
          onClick={() => createQueryTab(selectedContext || undefined)}
          title="新建 SQL 标签页"
        >
          <Plus className="h-4 w-4" />
        </button>
      </div>

      {/* SQL 编辑器（固定区域 + 可拖拽分割线） */}
      <div style={{ height: editorHeight }} className="shrink-0 flex flex-col overflow-hidden min-h-0">
        {activeTab && (
          <SqlEditor
            connID={activeTab.connID}
            connectionType={activeTab.connectionType}
            database={activeTab.database}
            schema={activeTab.schema}
            contextLabel={activeTab.contextLabel}
            sql={activeTab.sql}
            onSqlChange={updateActiveTabSql}
          />
        )}
      </div>

      {/* 拖拽分割线 */}
      <div
        className="h-1 shrink-0 cursor-row-resize hover:bg-primary/30 active:bg-primary/50 transition-colors flex items-center justify-center group"
        onMouseDown={onMouseDown}
      >
        <div className="w-8 h-0.5 rounded-full bg-border group-hover:bg-primary/50 transition-colors" />
      </div>

      {/* 结果区域：Tab 容器 */}
      <div className="flex-1 overflow-hidden flex flex-col min-h-0">
        {results.length === 0 ? (
          <div className="flex items-center justify-center h-full text-sm text-muted-foreground">
            执行 SQL 查询查看结果
          </div>
        ) : (
          <>
            {/* Tab 栏 */}
            <div className="flex items-center border-b border-border shrink-0 bg-muted/30 overflow-x-auto">
              {results.map((r, i) => (
                <Tooltip key={i} delayDuration={300}>
                  <ContextMenu>
                    <ContextMenuTrigger asChild>
                      <TooltipTrigger asChild>
                        <div
                          className={
                            'flex items-center gap-1 px-3 py-1.5 text-xs border-r border-border cursor-pointer whitespace-nowrap group ' +
                            (i === activeResultIndex
                              ? 'bg-background text-foreground border-b-2 border-b-primary'
                              : 'text-muted-foreground hover:bg-accent/30 hover:text-foreground')
                          }
                          onClick={() => setActiveResult(i)}
                        >
                          <span>查询 {i + 1}</span>
                          <span className="text-muted-foreground">({r.rows?.length || 0} 行)</span>
                          <button
                            className="ml-1 p-0.5 rounded hover:bg-accent/50 opacity-0 group-hover:opacity-100 transition-opacity text-foreground"
                            onClick={(e) => { e.stopPropagation(); closeResult(i); }}
                            title="关闭"
                          >
                            <X className="h-3 w-3" />
                          </button>
                        </div>
                      </TooltipTrigger>
                    </ContextMenuTrigger>
                    <ContextMenuContent>
                      <ContextMenuItem onClick={() => copySql(r.sql)}>
                        复制 SQL
                      </ContextMenuItem>
                      <ContextMenuSeparator />
                      <ContextMenuItem onClick={() => closeResult(i)}>
                        关闭
                      </ContextMenuItem>
                      <ContextMenuItem onClick={() => closeOtherResults(i)}>
                        关闭其他
                      </ContextMenuItem>
                      <ContextMenuItem onClick={() => closeAllResults()}>
                        关闭全部
                      </ContextMenuItem>
                    </ContextMenuContent>
                  </ContextMenu>
                  {r.sql && (
                    <TooltipContent side="bottom" align="start" className="max-w-[560px]">
                      <pre className="whitespace-pre-wrap break-all font-mono text-xs leading-relaxed max-h-[240px] overflow-auto">
                        {r.sql}
                      </pre>
                    </TooltipContent>
                  )}
                </Tooltip>
              ))}
            </div>

            {/* 当前结果面板 */}
            <div className="flex-1 min-h-0">
              {results[activeResultIndex] && (
                <ResultPanel result={results[activeResultIndex]} />
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default SqlWorkspace;
