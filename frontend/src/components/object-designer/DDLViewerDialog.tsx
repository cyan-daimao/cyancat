import React from 'react';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Loader2, Copy, Code2 } from 'lucide-react';
import { useDesignerStore } from '@/stores/designer';
import { schemaApi } from '@/lib/api/schema';
import { toast } from '@/components/ui/use-toast';

const DDLViewerDialog: React.FC = () => {
  const { ddlViewerOpen, ddlViewerContext, closeDDLViewer } = useDesignerStore();

  const [ddl, setDdl] = React.useState('');
  const [loading, setLoading] = React.useState(false);

  React.useEffect(() => {
    if (!ddlViewerOpen || !ddlViewerContext) return;
    let cancelled = false;
    setLoading(true);
    setDdl('');

    schemaApi.getCreateTableDDL(
      ddlViewerContext.connID,
      ddlViewerContext.database,
      ddlViewerContext.schema,
      ddlViewerContext.table,
    )
      .then(sql => { if (!cancelled) setDdl(sql || ''); })
      .catch(e => { if (!cancelled) toast({ title: '获取 DDL 失败', description: e.message, variant: 'destructive' }); })
      .finally(() => { if (!cancelled) setLoading(false); });

    return () => { cancelled = true; };
  }, [ddlViewerOpen, ddlViewerContext?.connID, ddlViewerContext?.database, ddlViewerContext?.schema, ddlViewerContext?.table]);

  const handleCopy = async () => {
    if (!ddl) return;
    await navigator.clipboard.writeText(ddl);
    toast({ title: 'DDL 已复制' });
  };

  const handleOpenInEditor = async () => {
    if (!ddl) return;
    await navigator.clipboard.writeText(ddl);
    toast({ title: 'DDL 已复制到剪贴板', description: '可粘贴到 SQL 编辑器中' });
  };

  const tableName = ddlViewerContext?.table || '';

  return (
    <Dialog open={ddlViewerOpen} onOpenChange={(o) => !o && closeDDLViewer()}>
      <DialogContent className="sm:max-w-[720px] max-h-[75vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            DDL 查看器
            {tableName && <span className="text-muted-foreground font-normal">— {tableName}</span>}
          </DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-6 w-6 animate-spin mr-2" />
            加载 DDL...
          </div>
        ) : (
          <div className="flex-1 bg-zinc-900 rounded-md p-4 overflow-auto min-h-[300px]">
            <pre className="text-xs font-mono text-green-400 whitespace-pre-wrap break-all leading-relaxed">
              {ddl || '-- 无 DDL 内容'}
            </pre>
          </div>
        )}

        <DialogFooter className="gap-2 shrink-0">
          <Button variant="ghost" onClick={handleCopy} disabled={!ddl || loading}>
            <Copy className="h-4 w-4 mr-1" />
            复制 DDL
          </Button>
          <Button variant="outline" onClick={handleOpenInEditor} disabled={!ddl}>
            <Code2 className="h-4 w-4 mr-1" />
            在编辑器中打开
          </Button>
          <Button variant="outline" onClick={() => closeDDLViewer()}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default DDLViewerDialog;