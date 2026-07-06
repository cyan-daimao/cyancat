import React from 'react';
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle,
  AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction,
} from '@/components/ui/alert-dialog';
import { AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useDesignerStore } from '@/stores/designer';
import { useConnectionStore } from '@/stores/connection';
import { useSchemaStore } from '@/stores/schema';
import { toast } from '@/components/ui/use-toast';

const DeleteConnectionConfirmDialog: React.FC = () => {
  const {
    deleteConnectionConfirmOpen,
    deleteConnectionContext,
    closeDeleteConnectionConfirm,
  } = useDesignerStore();
  const { deleteConnection } = useConnectionStore();
  const { resetTree } = useSchemaStore();

  const [confirmed, setConfirmed] = React.useState(false);
  const [deleting, setDeleting] = React.useState(false);

  React.useEffect(() => {
    if (deleteConnectionConfirmOpen) {
      setConfirmed(false);
      setDeleting(false);
    }
  }, [deleteConnectionConfirmOpen]);

  const handleDelete = async () => {
    if (!deleteConnectionContext) return;
    setDeleting(true);
    try {
      resetTree(deleteConnectionContext.connID);
      const ok = await deleteConnection(deleteConnectionContext.connID);
      if (ok) {
        closeDeleteConnectionConfirm();
      }
    } catch (e: any) {
      toast({ title: '删除连接失败', description: e.message, variant: 'destructive' });
    } finally {
      setDeleting(false);
    }
  };

  return (
    <AlertDialog open={deleteConnectionConfirmOpen} onOpenChange={(o) => !o && closeDeleteConnectionConfirm()}>
      <AlertDialogContent className="sm:max-w-[480px]">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="h-5 w-5" />
            确认删除连接
          </AlertDialogTitle>
          <AlertDialogDescription>
            即将永久删除连接 <strong>{deleteConnectionContext?.name}</strong>，此操作不可撤销。已保存的密码等连接信息将被清除。
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="space-y-3">
          <label className="flex items-start gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={confirmed}
              onChange={e => setConfirmed(e.target.checked)}
              className="mt-0.5"
            />
            <span>我了解此操作将永久删除该连接配置，确认执行</span>
          </label>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleting}>取消</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            disabled={!confirmed || deleting}
            className={cn(
              'bg-destructive text-destructive-foreground hover:bg-destructive/90',
              (!confirmed || deleting) && 'pointer-events-none opacity-50',
            )}
          >
            {deleting ? '删除中...' : '确认删除'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};

export default DeleteConnectionConfirmDialog;
