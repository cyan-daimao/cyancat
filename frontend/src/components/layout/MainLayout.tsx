import React from 'react';
import { Separator } from '@/components/ui/separator';
import { Button } from '@/components/ui/button';
import { Database, ChevronRight } from 'lucide-react';
import { useConnectionStore } from '@/stores/connection';
import type { ConnectionDTO } from '@/lib/api/types';
import { cn } from '@/lib/utils';
import ObjectTree from '@/components/object-tree/ObjectTree';
import SqlWorkspace from '@/components/sql-editor/SqlWorkspace';
import VerticalResizer from '@/components/layout/VerticalResizer';
import ConnectionDialog from '@/components/connection/ConnectionDialog';
import CreateDatabaseDialog from '@/components/object-designer/CreateDatabaseDialog';
import TableDesignerDialog from '@/components/object-designer/TableDesignerDialog';
import DDLViewerDialog from '@/components/object-designer/DDLViewerDialog';
import DropTableConfirmDialog from '@/components/object-designer/DropTableConfirmDialog';
import DropDatabaseConfirmDialog from '@/components/object-designer/DropDatabaseConfirmDialog';
import DeleteConnectionConfirmDialog from '@/components/object-designer/DeleteConnectionConfirmDialog';
import McpServerDialog from '@/components/mcp/McpServerDialog';

const DEFAULT_WIDTH = 256;
const MIN_WIDTH = 160;
const MAX_WIDTH = 480;
const COLLAPSED_WIDTH = 28;
const MIN_RIGHT_WIDTH = 320;
const STORAGE_KEY = 'mainLayout.leftPanel';

const clamp = (v: number, lo: number, hi: number) => Math.min(hi, Math.max(lo, v));

interface PersistedState {
  width?: number;
  collapsed?: boolean;
}

const CollapsedSidebarStrip: React.FC<{ onExpand: () => void }> = ({ onExpand }) => (
  <div className="h-full flex flex-col items-center pt-2">
    <Button
      variant="ghost"
      size="icon"
      className="h-6 w-6"
      onClick={onExpand}
      title="展开侧边栏"
    >
      <ChevronRight className="h-3.5 w-3.5" />
    </Button>
  </div>
);

const MainLayout: React.FC = () => {
  const [showConnectionDialog, setShowConnectionDialog] = React.useState(false);
  const [editConnection, setEditConnection] = React.useState<ConnectionDTO | null>(null);
  const fetchConnections = useConnectionStore(s => s.fetchConnections);

  const [width, setWidth] = React.useState<number>(DEFAULT_WIDTH);
  const [collapsed, setCollapsed] = React.useState<boolean>(false);
  const [isDragging, setIsDragging] = React.useState<boolean>(false);
  const widthAtDragStart = React.useRef<number>(DEFAULT_WIDTH);

  // 初始化连接列表
  React.useEffect(() => {
    fetchConnections();
  }, [fetchConnections]);

  // 从 localStorage 恢复
  React.useEffect(() => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return;
      const parsed = JSON.parse(raw) as PersistedState;
      if (typeof parsed.width === 'number') {
        setWidth(clamp(parsed.width, MIN_WIDTH, MAX_WIDTH));
      }
      if (typeof parsed.collapsed === 'boolean') {
        setCollapsed(parsed.collapsed);
      }
    } catch {
      /* ignore */
    }
  }, []);

  // 持久化（拖拽结束/状态变化时写入，拖拽过程中不写）
  React.useEffect(() => {
    if (isDragging) return;
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ width, collapsed }));
    } catch {
      /* ignore */
    }
  }, [width, collapsed, isDragging]);

  // 窗口缩窄时 clamp
  React.useEffect(() => {
    const onResize = () => {
      const maxAllowed = Math.min(MAX_WIDTH, window.innerWidth - MIN_RIGHT_WIDTH);
      setWidth(w => Math.min(w, Math.max(MIN_WIDTH, maxAllowed)));
    };
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

  // 快捷键：Cmd/Ctrl+B 切换收缩
  React.useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && (e.key === 'b' || e.key === 'B')) {
        e.preventDefault();
        setCollapsed(c => !c);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const handleResizeStart = React.useCallback(() => {
    widthAtDragStart.current = collapsed ? MIN_WIDTH : width;
    setIsDragging(true);
  }, [width, collapsed]);

  const handleResize = React.useCallback((deltaX: number) => {
    const next = clamp(widthAtDragStart.current + deltaX, MIN_WIDTH, MAX_WIDTH);
    setWidth(next);
    if (collapsed) setCollapsed(false);
  }, [collapsed]);

  const handleResizeEnd = React.useCallback(() => {
    setIsDragging(false);
  }, []);

  const handleDoubleClick = React.useCallback(() => {
    setWidth(DEFAULT_WIDTH);
    setCollapsed(false);
  }, []);

  const effectiveWidth = collapsed ? COLLAPSED_WIDTH : width;

  return (
    <div className="flex flex-col h-screen bg-background text-foreground">
      {/* 顶栏 */}
      <header className="flex items-center h-10 px-3 border-b border-border shrink-0">
        <div className="flex items-center gap-2 font-semibold text-sm">
          <Database className="h-4 w-4 text-primary" />
          <span>DBStudio</span>
        </div>
        <div className="flex-1" />
      </header>

      {/* 主体 */}
      <div className="flex flex-1 overflow-hidden">
        {/* 左侧面板 */}
        <aside
          style={{ width: effectiveWidth }}
          className={cn(
            'shrink-0 border-r border-border overflow-hidden bg-background',
            !isDragging && 'transition-[width] duration-150 ease-out',
          )}
        >
          {collapsed ? (
            <CollapsedSidebarStrip onExpand={() => setCollapsed(false)} />
          ) : (
            <ObjectTree
              onCreateConnection={() => {
                setEditConnection(null);
                setShowConnectionDialog(true);
              }}
              onShowProperties={(conn) => {
                setEditConnection(conn);
                setShowConnectionDialog(true);
              }}
              onCollapse={() => setCollapsed(true)}
            />
          )}
        </aside>

        {/* 拖拽分割线 */}
        <VerticalResizer
          onResize={handleResize}
          onResizeStart={handleResizeStart}
          onResizeEnd={handleResizeEnd}
          onDoubleClick={handleDoubleClick}
        />

        {/* 右侧工作区 */}
        <div className="flex-1 flex flex-col overflow-hidden">
          <SqlWorkspace />
        </div>
      </div>

      {/* 底栏 */}
      <footer className="flex items-center h-6 px-3 text-xs text-muted-foreground border-t border-border shrink-0">
        <span>DBStudio v0.1.0</span>
        <Separator orientation="vertical" className="mx-2 h-3" />
        <span>就绪</span>
      </footer>

      {/* 全局对话框（由 useDesignerStore 控制） */}
      <CreateDatabaseDialog />
      <TableDesignerDialog />
      <DDLViewerDialog />
      <DropTableConfirmDialog />
      <DropDatabaseConfirmDialog />
      <DeleteConnectionConfirmDialog />
      <McpServerDialog />
      <ConnectionDialog
        open={showConnectionDialog}
        onOpenChange={(open) => {
          setShowConnectionDialog(open);
          if (!open) setEditConnection(null);
        }}
        editConnection={editConnection}
      />
    </div>
  );
};

export default MainLayout;
