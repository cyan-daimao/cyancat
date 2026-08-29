import React from 'react';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Loader2, Table as TableIcon } from 'lucide-react';
import { useDesignerStore } from '@/stores/designer';
import { useConnectionStore } from '@/stores/connection';
import { schemaApi } from '@/lib/api/schema';
import { toast } from '@/components/ui/use-toast';
import { normalizeDataTypeFor, supportsTableOptions } from '@/lib/db-types';
import FieldGrid from './FieldGrid';
import IndexGrid from './IndexGrid';
import ForeignKeyGrid from './ForeignKeyGrid';
import TableOptionsForm from './TableOptionsForm';
import RiskConfirmDialog from './RiskConfirmDialog';
import type {
  ColumnSpecDTO, IndexSpecDTO, ForeignKeySpecDTO,
  ColumnDTO, IndexDTO, ForeignKeyDTO,
  CreateTableRequest, AlterTableRequest,
} from '@/lib/api/types';

// 内部字段草稿类型
export interface FieldDraft {
  id: string; // 临时 ID
  name: string;
  dataType: string;
  typeLength: number | null;
  precision: number | null;
  scale: number | null;
  nullable: boolean;
  autoIncrement: boolean;
  unsigned: boolean;
  defaultValue: string | null;
  comment: string;
  collation: string;
  // 变更追踪
  status: 'new' | 'existing' | 'modified' | 'deleted';
}

// 内部索引草稿类型
export interface IndexDraft {
  id: string;
  name: string;
  type: string; // PRIMARY / UNIQUE / NORMAL / FULLTEXT
  columns: string[];
  comment: string;
  status: 'new' | 'existing' | 'modified' | 'deleted';
}

// 内部外键草稿类型
export interface FKDraft {
  id: string;
  name: string;
  columns: string[];
  referencedSchema: string;
  referencedTable: string;
  referencedColumns: string[];
  onDelete: string;
  onUpdate: string;
  status: 'new' | 'existing' | 'modified' | 'deleted';
}

// 表选项草稿
export interface TableOptions {
  name: string;
  comment: string;
  engine: string;
  charset: string;
  collation: string;
}

let nextDraftId = 1;
function newDraftId(): string {
  return `draft_${nextDraftId++}`;
}

// ColumnDTO → FieldDraft
function columnToField(c: ColumnDTO, connectionType?: string): FieldDraft {
  return {
    id: newDraftId(),
    name: c.name,
    dataType: normalizeDataTypeFor(connectionType, c.databaseType),
    typeLength: c.typeLength,
    precision: c.precision,
    scale: c.scale,
    nullable: c.nullable,
    autoIncrement: c.autoIncrement,
    unsigned: c.unsigned,
    defaultValue: c.defaultValue,
    comment: c.comment,
    collation: c.collation,
    status: 'existing',
  };
}

// IndexDTO → IndexDraft
function indexToDraft(idx: IndexDTO): IndexDraft {
  return {
    id: newDraftId(),
    name: idx.name,
    type: idx.primary ? 'PRIMARY' : idx.unique ? 'UNIQUE' : 'NORMAL',
    columns: idx.columns || [],
    comment: idx.comment ?? '',
    status: 'existing',
  };
}

// ForeignKeyDTO → FKDraft
function fkToDraft(fk: ForeignKeyDTO): FKDraft {
  return {
    id: newDraftId(),
    name: fk.name,
    columns: fk.columns || [],
    referencedSchema: fk.referencedSchema,
    referencedTable: fk.referencedTable,
    referencedColumns: fk.referencedColumns || [],
    onDelete: fk.onDelete,
    onUpdate: fk.onUpdate,
    status: 'existing',
  };
}

// FieldDraft → ColumnSpecDTO
function fieldToSpec(f: FieldDraft): ColumnSpecDTO {
  return {
    name: f.name,
    dataType: f.dataType,
    typeLength: f.typeLength,
    precision: f.precision,
    scale: f.scale,
    nullable: f.nullable,
    autoIncrement: f.autoIncrement,
    unsigned: f.unsigned,
    defaultValue: f.defaultValue,
    comment: f.comment,
    collation: f.collation,
    first: false,
    after: '',
  };
}

// IndexDraft → IndexSpecDTO
function indexDraftToSpec(idx: IndexDraft): IndexSpecDTO {
  return {
    name: idx.name,
    type: idx.type,
    columns: idx.columns,
    comment: idx.comment,
  };
}

// FKDraft → ForeignKeySpecDTO
function fkDraftToSpec(fk: FKDraft): ForeignKeySpecDTO {
  return {
    name: fk.name,
    columns: fk.columns,
    referencedSchema: fk.referencedSchema,
    referencedTable: fk.referencedTable,
    referencedColumns: fk.referencedColumns,
    onDelete: fk.onDelete,
    onUpdate: fk.onUpdate,
  };
}

const TableDesignerDialog: React.FC = () => {
  const { tableDesignerOpen, tableDesignerContext, closeTableDesigner } = useDesignerStore();
  const connections = useConnectionStore(s => s.connections);

  const ctx = tableDesignerContext;
  const connectionType = connections.find(c => c.id === ctx?.connID)?.type;
  const isPG = connectionType === 'postgres';
  const isMy = supportsTableOptions(connectionType);
  const [fields, setFields] = React.useState<FieldDraft[]>([]);
  const [indexes, setIndexes] = React.useState<IndexDraft[]>([]);
  const [foreignKeys, setForeignKeys] = React.useState<FKDraft[]>([]);
  const [options, setOptions] = React.useState<TableOptions>({
    name: '', comment: '', engine: 'InnoDB', charset: 'utf8mb4', collation: 'utf8mb4_general_ci',
  });
  const [primaryKey, setPrimaryKey] = React.useState<string[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [saving, setSaving] = React.useState(false);
  const [showRiskConfirm, setShowRiskConfirm] = React.useState(false);
  const [activeTab, setActiveTab] = React.useState('fields');

  const isEdit = ctx?.mode === 'edit';
  const isView = ctx?.mode === 'view';

  // 打开时加载数据
  React.useEffect(() => {
    if (!tableDesignerOpen || !ctx) return;
    nextDraftId = 1;
    setActiveTab(ctx.focusTab || 'fields');

    if (isEdit && ctx.tableName) {
      setLoading(true);
      schemaApi.describeTable(ctx.connID, ctx.database, ctx.schema, ctx.tableName)
        .then(detail => {
          setFields((detail.columns ?? []).map(c => columnToField(c, connectionType)));
          setIndexes((detail.indexes ?? []).map(indexToDraft));
          setForeignKeys((detail.foreignKeys ?? []).map(fkToDraft));
          setOptions({
            name: detail.name,
            comment: detail.comment || '',
            engine: isMy ? 'InnoDB' : '',
            charset: isMy ? 'utf8mb4' : '',
            collation: isMy ? 'utf8mb4_general_ci' : '',
          });
          // 找出主键列
          const pk = (detail.indexes ?? []).find(i => i.primary);
          setPrimaryKey(pk?.columns || []);
        })
        .catch(e => toast({ title: '加载表结构失败', description: e.message, variant: 'destructive' }))
        .finally(() => setLoading(false));
    } else {
      // create 模式
      setFields([{
        id: newDraftId(), name: '', dataType: isPG ? 'INTEGER' : 'INT', typeLength: null,
        precision: null, scale: null, nullable: false, autoIncrement: false,
        unsigned: false, defaultValue: null, comment: '', collation: '',
        status: 'new',
      }]);
      setIndexes([]);
      setForeignKeys([]);
      setOptions({
        name: '', comment: '',
        engine: isMy ? 'InnoDB' : '',
        charset: isMy ? 'utf8mb4' : '',
        collation: isMy ? 'utf8mb4_general_ci' : '',
      });
      setPrimaryKey([]);
    }
  }, [tableDesignerOpen, ctx?.connID, ctx?.database, ctx?.schema, ctx?.tableName, ctx?.mode, connectionType]);

  // 检测破坏性操作
  const hasDestructiveOps = (): boolean => {
    return fields.some(f => f.status === 'deleted') ||
      indexes.some(i => i.status === 'deleted') ||
      foreignKeys.some(fk => fk.status === 'deleted');
  };

  const getDestructiveOps = (): string[] => {
    const ops: string[] = [];
    fields.filter(f => f.status === 'deleted').forEach(f => ops.push(`删除字段: ${f.name}`));
    indexes.filter(i => i.status === 'deleted').forEach(i => ops.push(`删除索引: ${i.name}`));
    foreignKeys.filter(fk => fk.status === 'deleted').forEach(fk => ops.push(`删除外键: ${fk.name}`));
    return ops;
  };

  const handleSave = async () => {
    if (!ctx) return;

    // 如果有破坏性操作，先弹出确认
    if (isEdit && hasDestructiveOps()) {
      setShowRiskConfirm(true);
      return;
    }

    await doSave();
  };

  const doSave = async () => {
    if (!ctx) return;
    setSaving(true);
    try {
      if (!isEdit) {
        const activeFields = fields.filter(f => f.status !== 'deleted');
        const activeIndexes = indexes.filter(i => i.status !== 'deleted');
        const activeFKs = foreignKeys.filter(fk => fk.status !== 'deleted');
        const req: CreateTableRequest = {
          connID: ctx.connID,
          database: ctx.database,
          schema: ctx.schema,
          name: options.name,
          columns: activeFields.map(fieldToSpec),
          primaryKey,
          indexes: activeIndexes.map(indexDraftToSpec),
          foreignKeys: activeFKs.map(fkDraftToSpec),
          engine: isMy ? options.engine : '',
          charset: isMy ? options.charset : '',
          collation: isMy ? options.collation : '',
          comment: options.comment,
        };
        await schemaApi.createTable(req);
        toast({ title: '表已创建', description: options.name });
      } else {
        const req: AlterTableRequest = {
          connID: ctx.connID,
          database: ctx.database,
          schema: ctx.schema,
          name: options.name,
          addColumns: fields.filter(f => f.status === 'new').map(fieldToSpec),
          dropColumns: fields.filter(f => f.status === 'deleted').map(f => f.name),
          renameColumns: [],
          modifyColumns: fields.filter(f => f.status === 'modified').map(fieldToSpec),
          addIndexes: indexes.filter(i => i.status === 'new').map(indexDraftToSpec),
          modifyIndexes: indexes.filter(i => i.status === 'modified').map(indexDraftToSpec),
          dropIndexes: indexes.filter(i => i.status === 'deleted').map(i => i.name),
          addForeignKeys: foreignKeys.filter(fk => fk.status === 'new').map(fkDraftToSpec),
          dropForeignKeys: foreignKeys.filter(fk => fk.status === 'deleted').map(fk => fk.name),
          engine: isMy ? options.engine : '',
          charset: isMy ? options.charset : '',
          collation: isMy ? options.collation : '',
          comment: options.comment,
        };
        await schemaApi.alterTable(req);
        toast({ title: '表已修改', description: options.name });
      }
      closeTableDesigner();
    } catch (e: any) {
      toast({ title: isEdit ? '修改表失败' : '创建表失败', description: e.message, variant: 'destructive' });
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <Dialog open={tableDesignerOpen} onOpenChange={(o) => !o && closeTableDesigner()}>
        <DialogContent className="sm:max-w-[960px] max-h-[85vh] flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <TableIcon className="h-5 w-5 text-green-600" />
              {isView ? '查看表结构' : isEdit ? '修改表' : '新建表'}
              {options.name && <span className="text-muted-foreground font-normal">— {options.name}</span>}
            </DialogTitle>
          </DialogHeader>

          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-6 w-6 animate-spin mr-2" />
              加载表结构...
            </div>
          ) : (
            <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 overflow-hidden">
              <TabsList>
                <TabsTrigger value="fields">字段</TabsTrigger>
                <TabsTrigger value="indexes">索引</TabsTrigger>
                <TabsTrigger value="foreignKeys">外键</TabsTrigger>
                <TabsTrigger value="options">选项</TabsTrigger>
              </TabsList>

              <div className="flex-1 overflow-auto mt-2">
                <TabsContent value="fields" className="mt-0">
                  <FieldGrid
                    fields={fields}
                    setFields={setFields}
                    primaryKey={primaryKey}
                    setPrimaryKey={setPrimaryKey}
                    connectionType={connectionType}
                    readOnly={isView}
                  />
                </TabsContent>

                <TabsContent value="indexes" className="mt-0">
                  <IndexGrid
                    indexes={indexes}
                    setIndexes={setIndexes}
                    fields={fields}
                    readOnly={isView}
                  />
                </TabsContent>

                <TabsContent value="foreignKeys" className="mt-0">
                  <ForeignKeyGrid
                    foreignKeys={foreignKeys}
                    setForeignKeys={setForeignKeys}
                    readOnly={isView}
                  />
                </TabsContent>

                <TabsContent value="options" className="mt-0">
                  <TableOptionsForm
                    options={options}
                    setOptions={setOptions}
                    connID={ctx?.connID || 0}
                    connectionType={connectionType}
                    readOnly={isView || isEdit}
                  />
                </TabsContent>
              </div>
            </Tabs>
          )}

          <DialogFooter className="gap-2 shrink-0">
            <Button variant="outline" onClick={() => closeTableDesigner()}>
              取消
            </Button>
            {!isView && (
              <>
                <Button onClick={handleSave} disabled={saving || (!isEdit && !options.name.trim())}>
                  {saving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
                  保存
                </Button>
              </>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <RiskConfirmDialog
        open={showRiskConfirm}
        onOpenChange={setShowRiskConfirm}
        destructiveOps={getDestructiveOps()}
        ddl=""
        onConfirm={() => {
          setShowRiskConfirm(false);
          doSave();
        }}
      />
    </>
  );
};

export default TableDesignerDialog;
