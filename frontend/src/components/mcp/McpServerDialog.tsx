import React from 'react';
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useDesignerStore } from '@/stores/designer';
import { mcpApi, McpPortConflictError } from '@/lib/api/mcp';
import { toast } from '@/components/ui/use-toast';
import { Bot, Copy, Power, PowerOff, Loader2 } from 'lucide-react';
import type { McpServerStatusDTO } from '@/lib/api/types';

const McpServerDialog: React.FC = () => {
  const { mcpServerDialogOpen, mcpServerDialogContext, closeMcpServerDialog } = useDesignerStore();
  const connID = mcpServerDialogContext?.connID ?? 0;

  const [status, setStatus] = React.useState<McpServerStatusDTO | null>(null);
  const [loading, setLoading] = React.useState(false);
  const [starting, setStarting] = React.useState(false);
  const [stopping, setStopping] = React.useState(false);

  const [allowSelect, setAllowSelect] = React.useState(true);
  const [allowInsert, setAllowInsert] = React.useState(false);
  const [allowUpdate, setAllowUpdate] = React.useState(false);
  const [allowDelete, setAllowDelete] = React.useState(false);
  const [allowDDL, setAllowDDL] = React.useState(false);

  const fetchStatus = React.useCallback(async () => {
    if (!connID) return;
    setLoading(true);
    try {
      const s = await mcpApi.getStatus(connID);
      setStatus(s);
      setAllowSelect(s.allowSelect);
      setAllowInsert(s.allowInsert);
      setAllowUpdate(s.allowUpdate);
      setAllowDelete(s.allowDelete);
      setAllowDDL(s.allowDDL);
    } catch (e: any) {
      toast({ title: '获取 MCP Server 状态失败', description: e.message, variant: 'destructive' });
    } finally {
      setLoading(false);
    }
  }, [connID]);

  React.useEffect(() => {
    if (mcpServerDialogOpen && connID) {
      fetchStatus();
    }
    if (!mcpServerDialogOpen) {
      setStatus(null);
    }
  }, [mcpServerDialogOpen, connID, fetchStatus]);

  const doStart = async (forceNewPort = false): Promise<void> => {
    if (!connID) return;
    setStarting(true);
    try {
      const s = await mcpApi.start({
        connID,
        allowSelect,
        allowInsert,
        allowUpdate,
        allowDelete,
        allowDDL,
        forceNewPort,
      });
      setStatus(s);
      toast({ title: 'MCP Server 已开启', description: s.address });
    } catch (e: any) {
      if (e instanceof McpPortConflictError) {
        const confirmed = window.confirm(e.message);
        if (confirmed) {
          return doStart(true);
        }
        return;
      }
      toast({ title: '开启 MCP Server 失败', description: e.message, variant: 'destructive' });
    } finally {
      setStarting(false);
    }
  };

  const handleStart = () => doStart(false);

  const handleStop = async () => {
    if (!connID) return;
    setStopping(true);
    try {
      await mcpApi.stop(connID);
      setStatus(prev => prev ? { ...prev, enabled: false, address: '', token: '' } : null);
      toast({ title: 'MCP Server 已关闭' });
    } catch (e: any) {
      toast({ title: '关闭 MCP Server 失败', description: e.message, variant: 'destructive' });
    } finally {
      setStopping(false);
    }
  };

  const copyText = async (text: string, title = '已复制') => {
    try {
      await navigator.clipboard.writeText(text);
      toast({ title, description: text });
    } catch (e: any) {
      toast({ title: '复制失败', description: e.message, variant: 'destructive' });
    }
  };

  const sseUrl = status?.enabled && status.address ? status.address : '';
  const token = status?.enabled && status.token ? status.token : '';

  const installCommand = sseUrl
    ? `claude mcp add --transport sse dbstudio "${sseUrl}" \\\n  --header "Authorization: Bearer ${token}"`
    : '';

  const uninstallCommand = 'claude mcp remove dbstudio';

  const curlExample = sseUrl
    ? `curl -N -H "Accept: text/event-stream" -H "Authorization: Bearer ${token}" "${sseUrl}"`
    : '';

  const clientConfig = sseUrl
    ? JSON.stringify({
        mcpServers: {
          dbstudio: {
            transport: 'sse',
            url: sseUrl,
            headers: {
              Authorization: `Bearer ${token}`,
            },
          },
        },
      }, null, 2)
    : '';

  const renderCheckbox = (
    id: string,
    label: string,
    checked: boolean,
    onChange: (v: boolean) => void,
    disabled?: boolean,
  ) => (
    <div className="flex items-center gap-2">
      <input
        id={id}
        type="checkbox"
        checked={checked}
        disabled={disabled}
        onChange={e => onChange(e.target.checked)}
        className="h-4 w-4 rounded border-border bg-background text-primary focus:ring-ring disabled:opacity-50"
      />
      <Label htmlFor={id} className="text-sm font-normal cursor-pointer">
        {label}
      </Label>
    </div>
  );

  const renderCopyRow = (label: string, value: string, copyTitle: string, password?: boolean) => (
    <div className="space-y-1.5">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      <div className="flex gap-2">
        <Input value={value} readOnly type={password ? 'password' : 'text'} className="font-mono text-xs" />
        <Button variant="outline" size="icon" onClick={() => copyText(value, copyTitle)}>
          <Copy className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );

  const renderCodeBlock = (title: string, code: string, copyTitle: string) => (
    <div>
      <div className="flex items-center justify-between mb-1">
        <span className="text-xs text-muted-foreground">{title}</span>
        <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={() => copyText(code, copyTitle)}>
          <Copy className="h-3 w-3 mr-1" />
          复制
        </Button>
      </div>
      <pre className="p-3 rounded-md bg-muted text-xs font-mono whitespace-pre-wrap break-all">{code}</pre>
    </div>
  );

  return (
    <Dialog open={mcpServerDialogOpen} onOpenChange={(o) => !o && closeMcpServerDialog()}>
      <DialogContent className="max-w-4xl" onPointerDownOutside={(e) => e.preventDefault()}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Bot className="h-5 w-5 text-primary" />
            MCP Server
          </DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center gap-2 text-sm text-muted-foreground py-8">
            <Loader2 className="h-4 w-4 animate-spin" />
            加载中…
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-2">
            {/* 左侧：权限与操作 */}
            <div className="space-y-5">
              <div>
                <div className="text-sm font-medium mb-2">允许执行的操作</div>
                <div className="grid grid-cols-2 gap-3 p-3 rounded-md border border-border bg-muted/30">
                  {renderCheckbox('mcp-allow-select', 'SELECT', allowSelect, setAllowSelect, starting || stopping)}
                  {renderCheckbox('mcp-allow-insert', 'INSERT', allowInsert, setAllowInsert, starting || stopping)}
                  {renderCheckbox('mcp-allow-update', 'UPDATE', allowUpdate, setAllowUpdate, starting || stopping)}
                  {renderCheckbox('mcp-allow-delete', 'DELETE', allowDelete, setAllowDelete, starting || stopping)}
                  {renderCheckbox('mcp-allow-ddl', 'DDL', allowDDL, setAllowDDL, starting || stopping)}
                </div>
              </div>

              {status?.enabled ? (
                <div className="flex items-center gap-2 text-sm text-green-700 dark:text-green-400 p-3 rounded-md border border-green-500/30 bg-green-500/10">
                  <span className="relative flex h-2.5 w-2.5">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75" />
                    <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-green-500" />
                  </span>
                  MCP Server 运行中
                </div>
              ) : (
                <div className="flex items-center gap-2 text-sm text-muted-foreground p-3 rounded-md border border-border bg-muted/30">
                  <PowerOff className="h-4 w-4" />
                  MCP Server 未开启
                </div>
              )}

              <div className="flex gap-2">
                {status?.enabled ? (
                  <Button variant="destructive" onClick={handleStop} disabled={stopping || loading} className="flex-1">
                    {stopping && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
                    <PowerOff className="h-4 w-4 mr-1" />
                    关闭 MCP Server
                  </Button>
                ) : (
                  <Button onClick={handleStart} disabled={starting || loading || !connID} className="flex-1">
                    {starting && <Loader2 className="h-4 w-4 mr-1 animate-spin" />}
                    <Power className="h-4 w-4 mr-1" />
                    开启 MCP Server
                  </Button>
                )}
              </div>
            </div>

            {/* 右侧：接入信息 */}
            <div className="space-y-4">
              {status?.enabled && sseUrl ? (
                <>
                  {renderCopyRow('SSE 地址', sseUrl, 'SSE 地址已复制')}
                  {renderCopyRow('访问 Token', token, 'Token 已复制', true)}
                  {renderCodeBlock('Claude Code 安装命令', installCommand, '安装命令已复制')}
                  {renderCodeBlock('Claude Code 卸载命令', uninstallCommand, '卸载命令已复制')}
                  {renderCodeBlock('curl 测试', curlExample, 'curl 命令已复制')}
                  {renderCodeBlock('MCP 客户端配置示例', clientConfig, '配置已复制')}
                </>
              ) : (
                <div className="h-full flex items-center justify-center text-sm text-muted-foreground border border-dashed border-border rounded-md p-6">
                  开启 MCP Server 后显示接入信息
                </div>
              )}
            </div>
          </div>
        )}

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => closeMcpServerDialog()}>
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default McpServerDialog;
