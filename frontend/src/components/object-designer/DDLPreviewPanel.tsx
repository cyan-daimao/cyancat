import React from 'react';
import { Button } from '@/components/ui/button';
import { Copy, Code2, ExternalLink } from 'lucide-react';
import { toast } from '@/components/ui/use-toast';

interface DDLPreviewPanelProps {
  ddl: string;
  onGenerate: () => void;
}

const DDLPreviewPanel: React.FC<DDLPreviewPanelProps> = ({ ddl, onGenerate }) => {
  const [copying, setCopying] = React.useState(false);

  const handleCopy = async () => {
    if (!ddl) return;
    setCopying(true);
    try {
      await navigator.clipboard.writeText(ddl);
      toast({ title: 'DDL 已复制' });
    } catch {
      toast({ title: '复制失败', variant: 'destructive' });
    } finally {
      setCopying(false);
    }
  };

  const handleOpenInEditor = async () => {
    if (!ddl) return;
    await navigator.clipboard.writeText(ddl);
    toast({ title: 'DDL 已复制到剪贴板', description: '可粘贴到 SQL 编辑器中' });
  };

  return (
    <div className="flex flex-col h-full min-h-[300px]">
      {/* 工具栏 */}
      <div className="flex items-center gap-2 pb-2">
        <Button size="sm" variant="outline" onClick={onGenerate}>
          <Code2 className="h-3 w-3 mr-1" />
          生成预览
        </Button>
        <Button size="sm" variant="ghost" onClick={handleCopy} disabled={!ddl || copying}>
          <Copy className="h-3 w-3 mr-1" />
          复制
        </Button>
        <Button size="sm" variant="ghost" onClick={handleOpenInEditor} disabled={!ddl}>
          <ExternalLink className="h-3 w-3 mr-1" />
          在编辑器中打开
        </Button>
      </div>

      {/* DDL 显示 */}
      {ddl ? (
        <div className="flex-1 bg-zinc-900 rounded-md p-4 overflow-auto">
          <pre className="text-xs font-mono text-green-400 whitespace-pre-wrap break-all leading-relaxed">
            {ddl}
          </pre>
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center bg-muted/30 rounded-md border border-dashed border-border">
          <div className="text-center text-muted-foreground">
            <Code2 className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">点击「生成预览」查看 DDL</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default DDLPreviewPanel;
