import React from 'react';
import { Separator } from '@/components/ui/separator';
import { Database } from 'lucide-react';
import { useConnectionStore } from '@/stores/connection';
import ObjectTree from '@/components/object-tree/ObjectTree';
import SqlWorkspace from '@/components/sql-editor/SqlWorkspace';
import CreateDatabaseDialog from '@/components/object-designer/CreateDatabaseDialog';
import TableDesignerDialog from '@/components/object-designer/TableDesignerDialog';
import DDLViewerDialog from '@/components/object-designer/DDLViewerDialog';
import DropTableConfirmDialog from '@/components/object-designer/DropTableConfirmDialog';

const MainLayout: React.FC = () => {
  const fetchConnections = useConnectionStore(s => s.fetchConnections);

  React.useEffect(() => {
    fetchConnections();
  }, [fetchConnections]);

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
        {/* 左侧面板 - 对象浏览器 */}
        <div className="w-64 border-r border-border flex flex-col shrink-0">
          <ObjectTree />
        </div>

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
    </div>
  );
};

export default MainLayout;
