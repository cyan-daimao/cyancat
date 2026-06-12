import React from 'react';
import { cn } from '@/lib/utils';
import type { FKDraft } from './TableDesignerDialog';

interface ForeignKeyGridProps {
  foreignKeys: FKDraft[];
  setForeignKeys: React.Dispatch<React.SetStateAction<FKDraft[]>>;
  readOnly?: boolean;
}

const ForeignKeyGrid: React.FC<ForeignKeyGridProps> = ({ foreignKeys, readOnly }) => {
  if (foreignKeys.length === 0) {
    return (
      <div className="border rounded-md p-6 text-center text-muted-foreground text-sm">
        暂无外键
      </div>
    );
  }

  return (
    <div className="border rounded-md overflow-hidden">
      <div className="overflow-auto max-h-[55vh]">
        <table className="w-full text-xs">
          <thead className="bg-muted sticky top-0 z-10">
            <tr className="text-left">
              <th className="px-2 py-1 w-10">#</th>
              <th className="px-2 py-1 min-w-[120px]">名称</th>
              <th className="px-2 py-1 min-w-[140px]">列</th>
              <th className="px-2 py-1 min-w-[140px]">引用表</th>
              <th className="px-2 py-1 min-w-[140px]">引用列</th>
              <th className="px-2 py-1 w-24">ON DELETE</th>
              <th className="px-2 py-1 w-24">ON UPDATE</th>
            </tr>
          </thead>
          <tbody>
            {foreignKeys.filter(fk => fk.status !== 'deleted').map((fk, i) => (
              <tr key={fk.id} className={cn(
                'border-b border-border/40',
                fk.status === 'new' && 'bg-green-500/5',
                fk.status === 'modified' && 'bg-amber-500/5',
              )}>
                <td className="px-2 py-0.5 text-muted-foreground">{i + 1}</td>
                <td className="px-2 py-0.5">{fk.name}</td>
                <td className="px-2 py-0.5 font-mono">{fk.columns.join(', ')}</td>
                <td className="px-2 py-0.5 font-mono">
                  {fk.referencedSchema && fk.referencedSchema !== '' ? `${fk.referencedSchema}.` : ''}
                  {fk.referencedTable}
                </td>
                <td className="px-2 py-0.5 font-mono">{fk.referencedColumns.join(', ')}</td>
                <td className="px-2 py-0.5">{fk.onDelete}</td>
                <td className="px-2 py-0.5">{fk.onUpdate}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {readOnly && (
        <div className="p-2 border-t bg-muted/30 text-xs text-muted-foreground">
          V1.0 仅支持查看外键，编辑功能将在后续版本实现
        </div>
      )}
    </div>
  );
};

export default ForeignKeyGrid;
