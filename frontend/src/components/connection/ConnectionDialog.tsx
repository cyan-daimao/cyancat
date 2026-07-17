import React from 'react';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogClose,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useConnectionStore } from '@/stores/connection';
import { FolderOpen, Loader2, Plug } from 'lucide-react';
import { fileDialogApi } from '@/lib/api/file-dialog';
import { toast } from '@/components/ui/use-toast';
import type { ConnectionDTO, CreateConnectionRequest } from '@/lib/api/types';

interface ConnectionDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editConnection?: ConnectionDTO | null;
}

const GROUPS = [
  { value: 'development', label: '开发' },
  { value: 'test', label: '测试' },
  { value: 'staging', label: '预发布' },
  { value: 'production', label: '生产' },
];

const emptyForm: CreateConnectionRequest = {
  name: '', type: 'mysql', host: '127.0.0.1', port: 3306,
  user: 'root', password: '', database: '', ssl: false,
  group: 'development', color: '',
};

const ConnectionDialog: React.FC<ConnectionDialogProps> = ({ open, onOpenChange, editConnection }) => {
  const { createConnection, updateConnection, testConnection } = useConnectionStore();
  const [form, setForm] = React.useState<CreateConnectionRequest>(emptyForm);
  const [testing, setTesting] = React.useState(false);
  const [saving, setSaving] = React.useState(false);
  const [selectingFile, setSelectingFile] = React.useState(false);

  React.useEffect(() => {
    if (editConnection) {
      setForm({
        name: editConnection.name,
        type: editConnection.type,
        host: editConnection.host,
        port: editConnection.port,
        user: editConnection.user,
        password: '',
        database: editConnection.database,
        ssl: editConnection.ssl,
        group: editConnection.group,
        color: editConnection.color,
      });
    } else {
      setForm(emptyForm);
    }
  }, [editConnection, open]);

  const handleTypeChange = (type: string) => {
    setForm(f => ({
      ...f,
      type,
      host: type === 'sqlite' ? '' : (f.host || '127.0.0.1'),
      port: type === 'mysql' ? 3306 : type === 'postgres' ? 5432 : type === 'starrocks' ? 9030 : 0,
      user: type === 'sqlite' ? '' : (f.user || 'root'),
      password: type === 'sqlite' ? '' : f.password,
      database: type === 'sqlite' ? '' : f.database,
      ssl: type === 'sqlite' ? false : f.ssl,
    }));
  };

  const databasePlaceholder = form.type === 'postgres' ? '留空默认 postgres' : '可选';
  const isSQLite = form.type === 'sqlite';
  const namePlaceholder = isSQLite ? '本地 SQLite' : '本地 MySQL';

  const validateForm = (forTest = false): boolean => {
    if (!form.name.trim() && !forTest) {
      toast({ title: '请输入连接名称', variant: 'destructive' });
      return false;
    }
    if (!form.host.trim()) {
      toast({
        title: isSQLite ? '请选择 SQLite 数据库文件' : '请输入主机地址',
        variant: 'destructive',
      });
      return false;
    }
    if (!isSQLite && !form.user.trim()) {
      toast({ title: '请输入用户名', variant: 'destructive' });
      return false;
    }
    return true;
  };

  const handleSelectSQLiteFile = async () => {
    setSelectingFile(true);
    try {
      const path = await fileDialogApi.selectSQLiteDatabaseFile();
      if (path) {
        setForm(f => ({
          ...f,
          host: path,
          name: f.name || path.split(/[\\/]/).pop()?.replace(/\.(sqlite3?|db)$/i, '') || 'SQLite',
        }));
      }
    } catch (e: any) {
      toast({ title: '选择文件失败', description: e.message, variant: 'destructive' });
    } finally {
      setSelectingFile(false);
    }
  };

  const handleTest = async () => {
    if (!validateForm(true)) return;
    setTesting(true);
    await testConnection({
      type: form.type, host: form.host, port: form.port,
      user: form.user, password: form.password, database: form.database, ssl: form.ssl,
    });
    setTesting(false);
  };

  const handleSave = async () => {
    if (!validateForm(false)) return;
    setSaving(true);
    // 保存失败（store 返回 null，错误已 toast）时保持对话框打开，避免表单内容丢失
    const saved = editConnection
      ? await updateConnection(editConnection.id, form)
      : await createConnection(form);
    setSaving(false);
    if (saved) {
      onOpenChange(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle>{editConnection ? '编辑连接' : '新建连接'}</DialogTitle>
        </DialogHeader>

        <Tabs defaultValue="basic" className="mt-2">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="basic">基本</TabsTrigger>
            <TabsTrigger value="advanced">高级</TabsTrigger>
          </TabsList>

          <TabsContent value="basic" className="space-y-3 mt-3">
            <div className="grid grid-cols-4 items-center gap-2">
              <Label className="text-right">名称</Label>
              <Input className="col-span-3" value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder={namePlaceholder} />
            </div>
            <div className="grid grid-cols-4 items-center gap-2">
              <Label className="text-right">类型</Label>
              <Select value={form.type} onValueChange={handleTypeChange}>
                <SelectTrigger className="col-span-3"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="mysql">MySQL</SelectItem>
                  <SelectItem value="postgres">PostgreSQL</SelectItem>
                  <SelectItem value="starrocks">StarRocks</SelectItem>
                  <SelectItem value="sqlite">SQLite</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {isSQLite ? (
              <div className="grid grid-cols-4 items-center gap-2">
                <Label className="text-right">数据库文件</Label>
                <div className="col-span-3 flex gap-2">
                  <Input
                    value={form.host}
                    onChange={e => setForm(f => ({ ...f, host: e.target.value }))}
                    placeholder="请选择已有 .sqlite / .db 文件"
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={handleSelectSQLiteFile}
                    disabled={selectingFile}
                    title="选择 SQLite 文件"
                  >
                    {selectingFile ? <Loader2 className="h-4 w-4 animate-spin" /> : <FolderOpen className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
            ) : (
              <>
                <div className="grid grid-cols-4 items-center gap-2">
                  <Label className="text-right">主机</Label>
                  <Input className="col-span-3" value={form.host} onChange={e => setForm(f => ({ ...f, host: e.target.value }))} />
                </div>
                <div className="grid grid-cols-4 items-center gap-2">
                  <Label className="text-right">端口</Label>
                  <Input type="number" className="col-span-3" value={form.port} onChange={e => setForm(f => ({ ...f, port: parseInt(e.target.value) || 0 }))} />
                </div>
                <div className="grid grid-cols-4 items-center gap-2">
                  <Label className="text-right">用户</Label>
                  <Input className="col-span-3" value={form.user} onChange={e => setForm(f => ({ ...f, user: e.target.value }))} />
                </div>
                <div className="grid grid-cols-4 items-center gap-2">
                  <Label className="text-right">密码</Label>
                  <Input type="password" className="col-span-3" value={form.password} onChange={e => setForm(f => ({ ...f, password: e.target.value }))} />
                </div>
                <div className="grid grid-cols-4 items-center gap-2">
                  <Label className="text-right">数据库</Label>
                  <Input className="col-span-3" value={form.database} onChange={e => setForm(f => ({ ...f, database: e.target.value }))} placeholder={databasePlaceholder} />
                </div>
              </>
            )}
          </TabsContent>

          <TabsContent value="advanced" className="space-y-3 mt-3">
            <div className="grid grid-cols-4 items-center gap-2">
              <Label className="text-right">分组</Label>
              <Select value={form.group} onValueChange={v => setForm(f => ({ ...f, group: v }))}>
                <SelectTrigger className="col-span-3"><SelectValue /></SelectTrigger>
                <SelectContent>
                  {GROUPS.map(g => <SelectItem key={g.value} value={g.value}>{g.label}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>
            {!isSQLite && (
              <div className="grid grid-cols-4 items-center gap-2">
                <Label className="text-right">SSL</Label>
                <div className="col-span-3 flex items-center gap-2">
                  <input type="checkbox" checked={form.ssl} onChange={e => setForm(f => ({ ...f, ssl: e.target.checked }))} />
                  <span className="text-sm text-muted-foreground">启用 SSL 连接</span>
                </div>
              </div>
            )}
          </TabsContent>
        </Tabs>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={handleTest} disabled={testing}>
            {testing ? <Loader2 className="h-4 w-4 mr-1 animate-spin" /> : <Plug className="h-4 w-4 mr-1" />}
            测试连接
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
            保存
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default ConnectionDialog;
