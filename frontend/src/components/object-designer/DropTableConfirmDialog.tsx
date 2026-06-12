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

const DropTableConfirmDialog: React.FC = () => {
  const { dropTableConfirmOpen, dropTableContext, closeDropTableConfirm } = useDesignerStore();
  const { invalidateDatabase } = useSchemaStore();

  const [confirmed, setConfirmed] = React.useState(false);
  const [ddl, setDdl] = React.useState('');
  const [loading, setLoading] = React.useState(false);
  const [dropping, setDropping] = React.useState(false);

  // 打开时预览 DDL
  React.useEffect(() => {
    if (dropTableConfirmOpen && dropTableContext) {
      setConfirmed(false);
      setDdl('');
      setLoading(true);
      schemaApi.previewDropTable({
        connID: dropTableContext.connID,
        database: dropTableContext.database,
        schema: dropTableContext.schema,
        name: dropTableContext.tableName,
      })
        .then((d: string) => setDdl(d))
        .catch((e: Error) => setDdl(`-- 预览失败: ${e.message}`))
        .finally(() => setLoading(false));
    }
  }, [dropTableConfirmOpen, dropTableContext]);

  const handleDrop = async () => {
    if (!dropTableContext) return;
    setDropping(true);
    try {
      await schemaApi.dropTable({
        connID: dropTableContext.connID,
        database: dropTableContext.database,
        schema: dropTableContext.schema,
        name: dropTableContext.tableName,
      });
      toast({ title: '表已删除', description: dropTableContext.tableName });
      closeDropTableConfirm();
      // 刷新对象树
      await invalidateDatabase(dropTableContext.connID, dropTableContext.database);
    } catch (e: any) {
      toast({ title: '删除表失败', description: e.message, variant: 'destructive' });
    } finally {
      setDropping(false);
    }
  };

  return (
    <AlertDialog open={dropTableConfirmOpen} onOpenChange={(o) => !o && closeDropTableConfirm()}>
      <AlertDialogContent className="sm:max-w-[560px]">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="h-5 w-5" />
            确认删除表
          </AlertDialogTitle>
          <AlertDialogDescription>
            即将永久删除表 <strong>{dropTableContext?.tableName}</strong>，此操作不可撤销！表中的所有数据将被永久丢失。
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="space-y-3">
          {/* DDL 预览 */}
          {loading ? (
            <div className="text-sm text-muted-foreground">加载 DDL 预览...</div>
          ) : ddl && (
            <div className="p-3 bg-muted rounded-md">
              <div className="text-xs text-muted-foreground mb-1">将要执行的 DDL：</div>
              <pre className="text-xs font-mono whitespace-pre-wrap break-all max-h-[200px] overflow-auto text-destructive">{ddl}</pre>
            </div>
          )}

          {/* 确认复选框 */}
          <label className="flex items-start gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={confirmed}
              onChange={e => setConfirmed(e.target.checked)}
              className="mt-0.5"
            />
            <span>我了解此操作将永久删除表及所有数据，确认执行</span>
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

export default DropTableConfirmDialog;
