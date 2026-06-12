import React from 'react';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useDesignerStore } from '@/stores/designer';
import { useSchemaStore } from '@/stores/schema';
import { schemaApi } from '@/lib/api/schema';
import { toast } from '@/components/ui/use-toast';
import { Loader2, Database, Code2 } from 'lucide-react';
import type { CharsetDTO, CollationDTO } from '@/lib/api/types';

const CreateDatabaseDialog: React.FC = () => {
  const { createDatabaseOpen, createDatabaseContext, closeCreateDatabase } = useDesignerStore();
  const { refreshTree } = useSchemaStore();

  const [name, setName] = React.useState('');
  const [charset, setCharset] = React.useState('utf8mb4');
  const [collation, setCollation] = React.useState('utf8mb4_general_ci');
  const [autoSelect, setAutoSelect] = React.useState(true);

  const [charsets, setCharsets] = React.useState<CharsetDTO[]>([]);
  const [collations, setCollations] = React.useState<CollationDTO[]>([]);
  const [loadingCharsets, setLoadingCharsets] = React.useState(false);

  const [previewSQL, setPreviewSQL] = React.useState('');
  const [showPreview, setShowPreview] = React.useState(false);
  const [saving, setSaving] = React.useState(false);

  const connID = createDatabaseContext?.connID;

  // 打开时加载字符集列表
  React.useEffect(() => {
    if (!createDatabaseOpen || !connID) return;
    setName('');
    setCharset('utf8mb4');
    setCollation('utf8mb4_general_ci');
    setAutoSelect(true);
    setShowPreview(false);
    setPreviewSQL('');

    setLoadingCharsets(true);
    schemaApi.listCharsets(connID)
      .then(list => setCharsets(list || []))
      .catch(() => setCharsets([]))
      .finally(() => setLoadingCharsets(false));
  }, [createDatabaseOpen, connID]);

  // 字符集变化时加载对应排序规则
  React.useEffect(() => {
    if (!connID || !charset) return;
    schemaApi.listCollations(connID, charset)
      .then(list => {
        setCollations(list || []);
        // 找默认排序规则
        const def = (list || []).find(c => c.isDefault);
        const cs = charsets.find(c => c.name === charset);
        if (def) {
          setCollation(def.name);
        } else if (cs?.defaultCollation) {
          setCollation(cs.defaultCollation);
        }
      })
      .catch(() => setCollations([]));
  }, [connID, charset, charsets]);

  // 简单本地预览（不调后端）
  const buildPreview = (): string => {
    if (!name.trim()) return '';
    let sql = `CREATE DATABASE \`${name.trim()}\``;
    if (charset) sql += ` DEFAULT CHARACTER SET ${charset}`;
    if (collation) sql += ` COLLATE ${collation}`;
    return sql + ';';
  };

  const handlePreview = () => {
    setPreviewSQL(buildPreview());
    setShowPreview(true);
  };

  const handleSave = async () => {
    if (!connID || !name.trim()) return;
    setSaving(true);
    try {
      await schemaApi.createDatabase({
        connID,
        name: name.trim(),
        charset,
        collation,
      });
      toast({ title: '数据库已创建', description: name });
      // 刷新对象树
      await refreshTree(connID);
      closeCreateDatabase();
    } catch (e: any) {
      toast({ title: '创建数据库失败', description: e.message, variant: 'destructive' });
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={createDatabaseOpen} onOpenChange={(o) => !o && closeCreateDatabase()}>
      <DialogContent className="sm:max-w-[560px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Database className="h-5 w-5 text-amber-500" />
            新建数据库
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-3 mt-2">
          <div className="grid grid-cols-4 items-center gap-2">
            <Label className="text-right">数据库名 *</Label>
            <Input
              className="col-span-3"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="例如：my_database"
              autoFocus
            />
          </div>

          <div className="grid grid-cols-4 items-center gap-2">
            <Label className="text-right">字符集</Label>
            <Select value={charset} onValueChange={setCharset}>
              <SelectTrigger className="col-span-3">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-[300px]">
                {loadingCharsets ? (
                  <SelectItem value="utf8mb4">utf8mb4</SelectItem>
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
            <Select value={collation} onValueChange={setCollation}>
              <SelectTrigger className="col-span-3">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="max-h-[300px]">
                {collations.length === 0 ? (
                  <SelectItem value="utf8mb4_general_ci">utf8mb4_general_ci</SelectItem>
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

          <div className="grid grid-cols-4 items-center gap-2">
            <Label className="text-right">创建后选中</Label>
            <div className="col-span-3 flex items-center gap-2">
              <input
                type="checkbox"
                checked={autoSelect}
                onChange={e => setAutoSelect(e.target.checked)}
              />
              <span className="text-sm text-muted-foreground">创建成功后自动选中该数据库</span>
            </div>
          </div>

          {showPreview && previewSQL && (
            <div className="mt-3 p-3 bg-muted rounded-md">
              <div className="text-xs text-muted-foreground mb-1 flex items-center gap-1">
                <Code2 className="h-3 w-3" />
                SQL 预览
              </div>
              <pre className="text-xs font-mono whitespace-pre-wrap break-all">{previewSQL}</pre>
            </div>
          )}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={handlePreview} disabled={!name.trim()}>
            <Code2 className="h-4 w-4 mr-1" />
            预览 SQL
          </Button>
          <Button variant="outline" onClick={() => closeCreateDatabase()}>
            取消
          </Button>
          <Button onClick={handleSave} disabled={saving || !name.trim()}>
            {saving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
            保存
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default CreateDatabaseDialog;
