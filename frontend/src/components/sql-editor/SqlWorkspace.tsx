import React from 'react';
import SqlEditor from './SqlEditor';
import ResultPanel from '@/components/data-table/ResultPanel';
import { useQueryStore } from '@/stores/query';
import { useSchemaStore } from '@/stores/schema';
import { X } from 'lucide-react';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';

const MIN_EDITOR_HEIGHT = 80;
const MAX_EDITOR_HEIGHT = 600;
const DEFAULT_EDITOR_HEIGHT = 200;

const SqlWorkspace: React.FC = () => {
  const { results, activeResultIndex, setActiveResult, closeResult, closeOtherResults, closeAllResults } = useQueryStore();
  const selectedNode = useSchemaStore(s => s.selectedNode);

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
      {/* SQL 编辑器（固定区域 + 可拖拽分割线） */}
      <div style={{ height: editorHeight }} className="shrink-0 flex flex-col overflow-hidden min-h-0">
        <SqlEditor
          connID={selectedNode?.connID || 0}
          database={selectedNode?.database}
          schema={selectedNode?.schema}
        />
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
                <ContextMenu key={i}>
                  <ContextMenuTrigger asChild>
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
                        className="ml-1 p-0.5 rounded hover:bg-accent/50 opacity-0 group-hover:opacity-100 transition-opacity"
                        onClick={(e) => { e.stopPropagation(); closeResult(i); }}
                        title="关闭"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </div>
                  </ContextMenuTrigger>
                  <ContextMenuContent>
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
