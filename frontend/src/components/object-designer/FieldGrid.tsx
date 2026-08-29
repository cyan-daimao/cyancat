import React from 'react';
import {
  ContextMenu,
  ContextMenuTrigger,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
} from '@/components/ui/context-menu';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, Trash2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { toast } from '@/components/ui/use-toast';
import { getCommonTypes, supportsUnsigned } from '@/lib/db-types';
import type { FieldDraft } from './TableDesignerDialog';

interface FieldGridProps {
  fields: FieldDraft[];
  setFields: React.Dispatch<React.SetStateAction<FieldDraft[]>>;
  primaryKey: string[];
  setPrimaryKey: React.Dispatch<React.SetStateAction<string[]>>;
  connectionType?: string;
  readOnly?: boolean;
}

let nextId = 10000;
function newId(): string {
  return `draft_${++nextId}`;
}

function emptyField(connectionType?: string): FieldDraft {
  const defaultType = connectionType === 'postgres' ? 'INTEGER' : 'INT';
  return {
    id: newId(), name: '', dataType: defaultType, typeLength: null,
    precision: null, scale: null, nullable: true, autoIncrement: false,
    unsigned: false, defaultValue: null, comment: '', collation: '',
    status: 'new',
  };
}

const FieldGrid: React.FC<FieldGridProps> = ({ fields, setFields, primaryKey, setPrimaryKey, connectionType, readOnly }) => {
  const commonTypes = getCommonTypes(connectionType);
  const showUnsigned = supportsUnsigned(connectionType);
  const updateField = (id: string, patch: Partial<FieldDraft>) => {
    setFields(prev => prev.map(f => {
      if (f.id !== id) return f;
      const newStatus: FieldDraft['status'] =
        f.status === 'existing' ? 'modified' : f.status;
      return { ...f, ...patch, status: newStatus };
    }));
  };

  const addField = () => {
    setFields(prev => [...prev, emptyField(connectionType)]);
  };

  const insertAbove = (id: string) => {
    setFields(prev => {
      const idx = prev.findIndex(f => f.id === id);
      if (idx < 0) return prev;
      const next = [...prev];
      next.splice(idx, 0, emptyField(connectionType));
      return next;
    });
  };

  const insertBelow = (id: string) => {
    setFields(prev => {
      const idx = prev.findIndex(f => f.id === id);
      if (idx < 0) return prev;
      const next = [...prev];
      next.splice(idx + 1, 0, emptyField(connectionType));
      return next;
    });
  };

  const deleteField = (id: string) => {
    setFields(prev => prev.map(f => {
      if (f.id !== id) return f;
      if (f.status === 'new') return null as any;
      return { ...f, status: 'deleted' as const };
    }).filter(Boolean));
  };

  const duplicateField = (id: string) => {
    setFields(prev => {
      const idx = prev.findIndex(f => f.id === id);
      if (idx < 0) return prev;
      const src = prev[idx];
      const copy: FieldDraft = { ...src, id: newId(), name: src.name + '_copy', status: 'new' };
      const next = [...prev];
      next.splice(idx + 1, 0, copy);
      return next;
    });
  };

  const togglePK = (name: string) => {
    if (!name) return;
    setPrimaryKey(prev => prev.includes(name) ? prev.filter(n => n !== name) : [...prev, name]);
    // PK implies NOT NULL
    setFields(prev => prev.map(f => {
      if (f.name !== name) return f;
      if (primaryKey.includes(name)) return f; // toggling off
      return { ...f, nullable: false, status: f.status === 'existing' ? 'modified' : f.status };
    }));
  };

  const validate = (): boolean => {
    const names = new Set<string>();
    for (const f of fields) {
      if (f.status === 'deleted') continue;
      if (!f.name.trim()) {
        toast({ title: '校验失败', description: '存在空字段名', variant: 'destructive' });
        return false;
      }
      if (names.has(f.name)) {
        toast({ title: '校验失败', description: `字段名重复: ${f.name}`, variant: 'destructive' });
        return false;
      }
      names.add(f.name);
    }
    return true;
  };

  // 暴露 validate 方法（不在本组件 UI 中直接使用，但保留以便后续扩展）
  void validate;

  return (
    <div className="border rounded-md overflow-hidden">
      <div className="overflow-auto max-h-[55vh]">
        <table className="w-full text-xs">
          <thead className="bg-muted sticky top-0 z-10">
            <tr className="text-left">
              <th className="px-2 py-1 w-10">#</th>
              <th className="px-2 py-1 min-w-[140px]">名称 *</th>
              <th className="px-2 py-1 min-w-[120px]">类型</th>
              <th className="px-2 py-1 w-20">长度</th>
              <th className="px-2 py-1 w-16">小数</th>
              <th className="px-2 py-1 w-12 text-center">空</th>
              <th className="px-2 py-1 w-12 text-center">PK</th>
              <th className="px-2 py-1 w-12 text-center">AI</th>
              {showUnsigned && <th className="px-2 py-1 w-16 text-center">无符号</th>}
              <th className="px-2 py-1 min-w-[100px]">默认值</th>
              <th className="px-2 py-1 min-w-[140px]">注释</th>
              {!readOnly && <th className="px-2 py-1 w-12"></th>}
            </tr>
          </thead>
          <tbody>
            {fields.map((f, i) => {
              const deleted = f.status === 'deleted';
              const isPK = primaryKey.includes(f.name);
              return (
                <ContextMenu key={f.id}>
                  <ContextMenuTrigger asChild>
                    <tr className={cn(
                      'border-b border-border/40 hover:bg-accent/30',
                      deleted && 'opacity-50 line-through',
                      f.status === 'new' && 'bg-green-500/5',
                      f.status === 'modified' && 'bg-amber-500/5',
                    )}>
                      <td className="px-2 py-0.5 text-muted-foreground">{i + 1}</td>
                      <td className="px-1 py-0.5">
                        <Input
                          className="h-6 text-xs px-1"
                          value={f.name}
                          onChange={e => updateField(f.id, { name: e.target.value })}
                          disabled={readOnly || deleted}
                        />
                      </td>
                      <td className="px-1 py-0.5">
                        <Select
                          value={f.dataType}
                          onValueChange={v => updateField(f.id, { dataType: v })}
                          disabled={readOnly || deleted}
                        >
                          <SelectTrigger className="h-6 text-xs px-1"><SelectValue /></SelectTrigger>
                          <SelectContent>
                            {commonTypes.map(t => <SelectItem key={t} value={t}>{t}</SelectItem>)}
                          </SelectContent>
                        </Select>
                      </td>
                      <td className="px-1 py-0.5">
                        <Input
                          type="number"
                          className="h-6 text-xs px-1"
                          value={f.typeLength ?? ''}
                          onChange={e => updateField(f.id, { typeLength: e.target.value ? parseInt(e.target.value) : null })}
                          disabled={readOnly || deleted}
                        />
                      </td>
                      <td className="px-1 py-0.5">
                        <Input
                          type="number"
                          className="h-6 text-xs px-1"
                          value={f.scale ?? ''}
                          onChange={e => updateField(f.id, { scale: e.target.value ? parseInt(e.target.value) : null })}
                          disabled={readOnly || deleted}
                        />
                      </td>
                      <td className="px-2 py-0.5 text-center">
                        <input
                          type="checkbox"
                          checked={f.nullable}
                          onChange={e => updateField(f.id, { nullable: e.target.checked })}
                          disabled={readOnly || deleted || isPK}
                        />
                      </td>
                      <td className="px-2 py-0.5 text-center">
                        <input
                          type="checkbox"
                          checked={isPK}
                          onChange={() => togglePK(f.name)}
                          disabled={readOnly || deleted || !f.name}
                        />
                      </td>
                      <td className="px-2 py-0.5 text-center">
                        <input
                          type="checkbox"
                          checked={f.autoIncrement}
                          onChange={e => updateField(f.id, { autoIncrement: e.target.checked })}
                          disabled={readOnly || deleted}
                        />
                      </td>
                      {showUnsigned && (
                        <td className="px-2 py-0.5 text-center">
                          <input
                            type="checkbox"
                            checked={f.unsigned}
                            onChange={e => updateField(f.id, { unsigned: e.target.checked })}
                            disabled={readOnly || deleted}
                          />
                        </td>
                      )}
                      <td className="px-1 py-0.5">
                        <Input
                          className="h-6 text-xs px-1"
                          value={f.defaultValue ?? ''}
                          onChange={e => updateField(f.id, { defaultValue: e.target.value || null })}
                          disabled={readOnly || deleted}
                        />
                      </td>
                      <td className="px-1 py-0.5">
                        <Input
                          className="h-6 text-xs px-1"
                          value={f.comment}
                          onChange={e => updateField(f.id, { comment: e.target.value })}
                          disabled={readOnly || deleted}
                        />
                      </td>
                      {!readOnly && (
                        <td className="px-1 py-0.5 text-center">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 w-6 p-0 text-destructive"
                            onClick={() => deleteField(f.id)}
                            disabled={deleted}
                          >
                            <Trash2 className="h-3 w-3" />
                          </Button>
                        </td>
                      )}
                    </tr>
                  </ContextMenuTrigger>
                  {!readOnly && (
                    <ContextMenuContent>
                      <ContextMenuItem onClick={() => insertAbove(f.id)}>上方插入</ContextMenuItem>
                      <ContextMenuItem onClick={() => insertBelow(f.id)}>下方插入</ContextMenuItem>
                      <ContextMenuItem onClick={() => duplicateField(f.id)}>复制行</ContextMenuItem>
                      <ContextMenuSeparator />
                      <ContextMenuItem onClick={() => togglePK(f.name)} disabled={!f.name}>
                        {isPK ? '取消主键' : '设为主键'}
                      </ContextMenuItem>
                      <ContextMenuSeparator />
                      <ContextMenuItem
                        onClick={() => deleteField(f.id)}
                        className="text-destructive focus:text-destructive"
                      >
                        删除字段
                      </ContextMenuItem>
                    </ContextMenuContent>
                  )}
                </ContextMenu>
              );
            })}
          </tbody>
        </table>
      </div>
      {!readOnly && (
        <div className="p-2 border-t bg-muted/30">
          <Button size="sm" variant="outline" onClick={addField}>
            <Plus className="h-3 w-3 mr-1" />
            添加字段
          </Button>
          <span className="ml-3 text-xs text-muted-foreground">
            右键行可插入 / 删除 / 设置主键
          </span>
        </div>
      )}
    </div>
  );
};

export default FieldGrid;
