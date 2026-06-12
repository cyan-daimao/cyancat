import React from 'react';
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle,
  AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction,
} from '@/components/ui/alert-dialog';
import { AlertTriangle } from 'lucide-react';
import { cn } from '@/lib/utils';

interface RiskConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  destructiveOps: string[];
  ddl: string;
  onConfirm: () => void;
}

const RiskConfirmDialog: React.FC<RiskConfirmDialogProps> = ({
  open, onOpenChange, destructiveOps, ddl, onConfirm,
}) => {
  const [confirmed, setConfirmed] = React.useState(false);

  React.useEffect(() => {
    if (open) setConfirmed(false);
  }, [open]);

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent className="sm:max-w-[560px]">
        <AlertDialogHeader>
          <AlertDialogTitle className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="h-5 w-5" />
            破坏性操作确认
          </AlertDialogTitle>
          <AlertDialogDescription>
            以下操作将不可逆地修改数据库结构，请确认：
          </AlertDialogDescription>
        </AlertDialogHeader>

        <div className="space-y-3">
          {/* 破坏性操作列表 */}
          <div className="space-y-1">
            {destructiveOps.map((op, i) => (
              <div key={i} className="flex items-start gap-2 text-sm">
                <span className="text-destructive mt-0.5">•</span>
                <span className="text-destructive">{op}</span>
              </div>
            ))}
          </div>

          {/* DDL 预览 */}
          {ddl && (
            <div className="mt-3 p-3 bg-muted rounded-md">
              <div className="text-xs text-muted-foreground mb-1">将要执行的 DDL：</div>
              <pre className="text-xs font-mono whitespace-pre-wrap break-all max-h-[200px] overflow-auto">{ddl}</pre>
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
            <span>我了解上述操作的风险，确认执行</span>
          </label>
        </div>

        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            disabled={!confirmed}
            className={cn(!confirmed && 'pointer-events-none opacity-50')}
          >
            确认执行
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};

export default RiskConfirmDialog;
