import React from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { IndexDraft, FieldDraft } from './TableDesignerDialog';

const INDEX_TYPES = ['PRIMARY', 'UNIQUE', 'NORMAL', 'FULLTEXT'];

let nextId = 20000;
function newId(): string { return `idx_${++nextId}`; }

function emptyIndex(): IndexDraft {
  return { id: newId(), name: '', type: 'NORMAL', columns: [], comment: '', status: 'new' };
}

interface IndexGridProps {
  indexes: IndexDraft[];
  setIndexes: React.Dispatch<React.SetStateAction<IndexDraft[]>>;
  fields: FieldDraft[];
  readOnly?: boolean;
}

const IndexGrid: React.FC<IndexGridProps> = ({ indexes, setIndexes, fields, readOnly }) => {
  const availableColumns = fields.filter(f => f.status !== 'deleted' && f.name).map(f => f.name);

  const updateIndex = (id: string, patch: Partial<IndexDraft>) => {
    setIndexes(prev => prev.map(i => {
      if (i.id !== id) return i;
      const status: IndexDraft['status'] = i.status === 'existing' ? 'modified' : i.status;
      return { ...i, ...patch, status };
    }));
  };

  const addIndex = () => {
    setIndexes(prev => [...prev, emptyIndex()]);
  };

  const deleteIndex = (id: string) => {
    setIndexes(prev => prev.map(i => {
      if (i.id !== id) return i;
      if (i.status === 'new') return null as any;
      return { ...i, status: 'deleted' as const };
    }).filter(Boolean));
  };

  return (
    <div className="border rounded-md overflow-hidden">
      <div className="overflow-auto max-h-[55vh]">
        <table className="w-full text-xs">
          <thead className="bg-muted sticky top-0 z-10">
            <tr className="text-left">
              <th className="px-2 py-1 w-10">#</th>
              <th className="px-2 py-1 min-w-[140px]">名称</th>
              <th className="px-2 py-1 w-32">类型</th>
              <th className="px-2 py-1 min-w-[200px]">列</th>
              <th className="px-2 py-1 min-w-[140px]">注释</th>
              {!readOnly && <th className="px-2 py-1 w-12"></th>}
            </tr>
          </thead>
          <tbody>
            {indexes.length === 0 && (
              <tr>
                <td colSpan={readOnly ? 5 : 6} className="px-3 py-6 text-center text-muted-foreground">
                  暂无索引
                </td>
              </tr>
            )}
            {indexes.map((idx, i) => {
              const deleted = idx.status === 'deleted';
              return (
                <tr key={idx.id} className={cn(
                  'border-b border-border/40',
                  deleted && 'opacity-50 line-through',
                  idx.status === 'new' && 'bg-green-500/5',
                  idx.status === 'modified' && 'bg-amber-500/5',
                )}>
                  <td className="px-2 py-0.5 text-muted-foreground">{i + 1}</td>
                  <td className="px-1 py-0.5">
                    <Input
                      className="h-6 text-xs px-1"
                      value={idx.name}
                      onChange={e => updateIndex(idx.id, { name: e.target.value })}
                      disabled={readOnly || deleted}
                    />
                  </td>
                  <td className="px-1 py-0.5">
                    <Select
                      value={idx.type}
                      onValueChange={v => updateIndex(idx.id, { type: v })}
                      disabled={readOnly || deleted}
                    >
                      <SelectTrigger className="h-6 text-xs px-1"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {INDEX_TYPES.map(t => <SelectItem key={t} value={t}>{t}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </td>
                  <td className="px-1 py-0.5">
                    <ColumnMultiSelect
                      selected={idx.columns}
                      available={availableColumns}
                      onChange={cols => updateIndex(idx.id, { columns: cols })}
                      disabled={readOnly || deleted}
                    />
                  </td>
                  <td className="px-1 py-0.5">
                    <Input
                      className="h-6 text-xs px-1"
                      value={idx.comment}
                      onChange={e => updateIndex(idx.id, { comment: e.target.value })}
                      disabled={readOnly || deleted}
                    />
                  </td>
                  {!readOnly && (
                    <td className="px-1 py-0.5 text-center">
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 w-6 p-0 text-destructive"
                        onClick={() => deleteIndex(idx.id)}
                        disabled={deleted}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </td>
                  )}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
      {!readOnly && (
        <div className="p-2 border-t bg-muted/30">
          <Button size="sm" variant="outline" onClick={addIndex}>
            <Plus className="h-3 w-3 mr-1" />
            添加索引
          </Button>
        </div>
      )}
    </div>
  );
};

// 多选列下拉
const ColumnMultiSelect: React.FC<{
  selected: string[];
  available: string[];
  onChange: (cols: string[]) => void;
  disabled?: boolean;
}> = ({ selected, available, onChange, disabled }) => {
  const [open, setOpen] = React.useState(false);

  const toggle = (col: string) => {
    if (selected.includes(col)) onChange(selected.filter(c => c !== col));
    else onChange([...selected, col]);
  };

  return (
    <div className="relative">
      <button
        type="button"
        className="flex items-center gap-1 text-xs px-2 py-0.5 rounded border border-border hover:bg-accent/30 min-h-[24px] w-full text-left disabled:opacity-50"
        onClick={() => !disabled && setOpen(!open)}
        disabled={disabled}
      >
        {selected.length === 0 ? (
          <span className="text-muted-foreground">选择列...</span>
        ) : (
          <span className="truncate">{selected.join(', ')}</span>
        )}
      </button>
      {open && !disabled && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div className="absolute left-0 top-full z-50 mt-1 bg-popover border rounded-md shadow-lg p-1 min-w-[160px] max-h-[200px] overflow-auto">
            {available.length === 0 ? (
              <div className="text-xs text-muted-foreground px-2 py-1">请先添加字段</div>
            ) : (
              available.map(col => (
                <label
                  key={col}
                  className="flex items-center gap-2 px-2 py-1 text-xs hover:bg-accent/30 rounded cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={selected.includes(col)}
                    onChange={() => toggle(col)}
                    className="h-3 w-3"
                  />
                  {col}
                </label>
              ))
            )}
          </div>
        </>
      )}
    </div>
  );
};

export default IndexGrid;
