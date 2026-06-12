import React from 'react';
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle,
  AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction,
} from '@/components/ui/alert-dialog';
import { AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useDesignerStore } from '@/stores/designer';
import { useSchemaStore } from '@/stores/schema';
import { schemaApi } from '@/lib/api/schema';
import { toast } from '@/components/ui/use-toast';

const DropDatabaseConfirmDialog: React.FC = () => {
  const {
    dropDatabaseConfirmOpen,
    dropDatabaseContext,
    closeDropDatabaseConfirm,
  } = useDesignerStore();
  const { refreshTree } = useSchemaStore();

  const [confirmed, setConfirmed] = React.useState(false);
  const [ddl, setDdl] = React.useState('');
  const [loading, setLoading] = React.useState(false);
  const [dropping, setDropping] = React.useState(false);

  React.useEffect(() => {
    if (!dropDatabaseConfirmOpen || !dropDatabaseContext) return;
    setConfirmed(false);
    setDdl('');
    setLoading(true);
    schemaApi.previewDropDatabase({
      connID: dropDatabaseContext.connID,
      name: dropDatabaseContext.name,
    })
      .then((d: string) => setDdl(d))
      .catch((e: Error) => setDdl(`-- 预览失败: ${e.message}`))
      .finally(() => setLoading(false));
  }, [dropDatabaseConfirmOpen, dropDatabaseContext]);

  const handleDrop = async () => {
    if (!dropDatabaseContext) return;
    setDropping(true);
    try {
      await schemaApi.dropDatabase({
        connID: dropDatabaseContext.connID,
        name: dropDatabaseContext.name,
      });
      toast({ title: '数据库已删除', description: dropDatabaseContext.name });
      closeDropDatabaseConfirm();
      await refreshTree(dropDatabaseContext.connID);
    } catch (e: any) {
      toast({ title: '删除数据库失败', description: e.message, variant: 'destructive' });
    } finally {
      setDropping(false);
    }
  };

  return (
    <AlertDialog open={dropDatabaseConfirmOpen} onOpenChange={(o) => !o && closeDropDatabaseConfirm()}>
      <AlertDialogContent className="sm:max-w-[560px]">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="h-5 w-5" />
            确认删除数据库
          </AlertDialogTitle>
          <AlertDialogDescription>
            即将永久删除数据库 <strong>{dropDatabaseContext?.name}</strong>，此操作不可撤销，库中的所有对象和数据都将被删除。
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="space-y-3">
          {loading ? (
            <div className="text-sm text-muted-foreground">加载 DDL 预览...</div>
          ) : ddl && (
            <div className="p-3 bg-muted rounded-md">
              <div className="text-xs text-muted-foreground mb-1">将要执行的 DDL：</div>
              <pre className="text-xs font-mono whitespace-pre-wrap break-all max-h-[200px] overflow-auto text-destructive">{ddl}</pre>
            </div>
          )}

          <label className="flex items-start gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={confirmed}
              onChange={e => setConfirmed(e.target.checked)}
              className="mt-0.5"
            />
            <span>我了解此操作将永久删除数据库及所有数据，确认执行</span>
          </label>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={dropping}>取消</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDrop}
            disabled={!confirmed || dropping}
            className={cn(
              'bg-destructive text-destructive-foreground hover:bg-destructive/90',
              (!confirmed || dropping) && 'pointer-events-none opacity-50',
            )}
          >
            {dropping ? '删除中...' : '确认删除'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};

export default DropDatabaseConfirmDialog;
