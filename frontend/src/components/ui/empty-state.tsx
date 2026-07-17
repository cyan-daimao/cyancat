import React from 'react';
import type { LucideIcon } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description?: string;
  /** 可选的主操作按钮 */
  action?: { label: string; onClick: () => void };
  className?: string;
}

/** 通用空状态占位：图标 + 标题 + 描述 + 可选操作 */
const EmptyState: React.FC<EmptyStateProps> = ({ icon: Icon, title, description, action, className }) => (
  <div className={cn('flex flex-col items-center justify-center gap-1.5 text-center', className)}>
    <Icon className="h-8 w-8 text-muted-foreground/40" />
    <div className="text-sm text-muted-foreground">{title}</div>
    {description && <div className="text-xs text-muted-foreground/60">{description}</div>}
    {action && (
      <Button variant="outline" size="sm" className="mt-2 h-7 text-xs" onClick={action.onClick}>
        {action.label}
      </Button>
    )}
  </div>
);

export default EmptyState;
