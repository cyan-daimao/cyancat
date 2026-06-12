import React from 'react';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { schemaApi } from '@/lib/api/schema';
import type { CharsetDTO, CollationDTO } from '@/lib/api/types';
import type { TableOptions } from './TableDesignerDialog';

const ENGINES = ['InnoDB', 'MyISAM', 'Memory', 'Archive'];

interface TableOptionsFormProps {
  options: TableOptions;
  setOptions: React.Dispatch<React.SetStateAction<TableOptions>>;
  connID: number;
  readOnly?: boolean;
}

const TableOptionsForm: React.FC<TableOptionsFormProps> = ({ options, setOptions, connID, readOnly }) => {
  const [charsets, setCharsets] = React.useState<CharsetDTO[]>([]);
  const [collations, setCollations] = React.useState<CollationDTO[]>([]);

  React.useEffect(() => {
    if (!connID) return;
    schemaApi.listCharsets(connID).then(list => setCharsets(list || [])).catch(() => setCharsets([]));
  }, [connID]);

  React.useEffect(() => {
    if (!connID || !options.charset) return;
    schemaApi.listCollations(connID, options.charset)
      .then(list => setCollations(list || []))
      .catch(() => setCollations([]));
  }, [connID, options.charset]);

  const update = (patch: Partial<TableOptions>) => {
    setOptions(prev => ({ ...prev, ...patch }));
  };

  return (
    <div className="space-y-3 p-2">
      <div className="grid grid-cols-4 items-center gap-2">
        <Label className="text-right">表名 *</Label>
        <Input
          className="col-span-3"
          value={options.name}
          onChange={e => update({ name: e.target.value })}
          disabled={readOnly}
          placeholder="例如：users"
        />
      </div>

      <div className="grid grid-cols-4 items-center gap-2">
        <Label className="text-right">注释</Label>
        <Input
          className="col-span-3"
          value={options.comment}
          onChange={e => update({ comment: e.target.value })}
          disabled={readOnly}
          placeholder="表注释（可选）"
        />
      </div>

      <div className="grid grid-cols-4 items-center gap-2">
        <Label className="text-right">引擎</Label>
        <Select value={options.engine} onValueChange={v => update({ engine: v })} disabled={readOnly}>
          <SelectTrigger className="col-span-3"><SelectValue /></SelectTrigger>
          <SelectContent>
            {ENGINES.map(e => <SelectItem key={e} value={e}>{e}</SelectItem>)}
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-4 items-center gap-2">
        <Label className="text-right">字符集</Label>
        <Select value={options.charset} onValueChange={v => update({ charset: v })} disabled={readOnly}>
          <SelectTrigger className="col-span-3"><SelectValue /></SelectTrigger>
          <SelectContent className="max-h-[300px]">
            {charsets.length === 0 ? (
              <SelectItem value={options.charset || 'utf8mb4'}>{options.charset || 'utf8mb4'}</SelectItem>
            ) : (
              charsets.map(c => (
                <SelectItem key={c.name} value={c.name}>
                  {c.name} {c.description && <span className="text-muted-foreground text-xs">({c.description})</span>}
                </SelectItem>
              ))
            )}
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-4 items-center gap-2">
        <Label className="text-right">排序规则</Label>
        <Select value={options.collation} onValueChange={v => update({ collation: v })} disabled={readOnly}>
          <SelectTrigger className="col-span-3"><SelectValue /></SelectTrigger>
          <SelectContent className="max-h-[300px]">
            {collations.length === 0 ? (
              <SelectItem value={options.collation || 'utf8mb4_general_ci'}>
                {options.collation || 'utf8mb4_general_ci'}
              </SelectItem>
            ) : (
              collations.map(c => (
                <SelectItem key={c.name} value={c.name}>
                  {c.name} {c.isDefault && <span className="text-muted-foreground text-xs">(默认)</span>}
                </SelectItem>
              ))
            )}
          </SelectContent>
        </Select>
      </div>

      {readOnly && (
        <div className="text-xs text-muted-foreground italic text-center mt-4">
          编辑模式下表名不可修改
        </div>
      )}
    </div>
  );
};

export default TableOptionsForm;
