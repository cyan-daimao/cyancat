import React from 'react';
import Editor from '@monaco-editor/react';
import { Button } from '@/components/ui/button';
import { Play, Square } from 'lucide-react';
import { useQueryStore } from '@/stores/query';
import { toast } from '@/components/ui/use-toast';

interface SqlEditorProps {
  connID: number;
  database?: string;
  schema?: string;
}

const SqlEditor: React.FC<SqlEditorProps> = ({ connID, database, schema }) => {
  const [sql, setSql] = React.useState('SELECT 1;');
  const { execute, executing } = useQueryStore();
  const editorRef = React.useRef<any>(null);
  const contextRef = React.useRef({ connID, database, schema });

  React.useEffect(() => {
    contextRef.current = { connID, database, schema };
  }, [connID, database, schema]);

  const handleExecute = async () => {
    const ctx = contextRef.current;
    if (!ctx.connID) {
      toast({ title: '请先在左侧选择连接、数据库或表', variant: 'destructive' });
      return;
    }

    // 取选中部分或全部
    let sqlToRun = sql;
    if (editorRef.current) {
      const selection = editorRef.current.getSelection();
      const selectedText = editorRef.current.getModel().getValueInRange(selection);
      if (selectedText.trim()) {
        sqlToRun = selectedText;
      }
    }

    if (!sqlToRun.trim()) return;

    await execute({
      connID: ctx.connID,
      sql: sqlToRun,
      maxRows: 1000,
      database: ctx.database,
      schema: ctx.schema,
    });
  };

  // 用 ref 暴露 handleExecute 给 Monaco 命令
  const handleExecuteRef = React.useRef(handleExecute);
  React.useEffect(() => {
    handleExecuteRef.current = handleExecute;
  });

  const handleEditorMount = (editor: any, monaco: any) => {
    editorRef.current = editor;
    // Cmd/Ctrl + Enter 执行
    editor.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter,
      () => handleExecuteRef.current()
    );
  };

  return (
    <div className="flex flex-col h-full">
      {/* 工具栏 */}
      <div className="flex items-center gap-1 px-2 py-1 border-b border-border shrink-0">
        <Button size="sm" variant="ghost" onClick={handleExecute} disabled={executing || !connID}>
          {executing ? <Square className="h-3 w-3 mr-1" /> : <Play className="h-3 w-3 mr-1" />}
          {executing ? '执行中...' : '执行'}
        </Button>
        <span className="text-xs text-muted-foreground ml-2">⌘+Enter 执行选中 / 全部</span>
      </div>

      {/* 编辑器 */}
      <div className="flex-1">
        <Editor
          height="100%"
          language="sql"
          theme="vs-dark"
          value={sql}
          onChange={v => setSql(v || '')}
          onMount={handleEditorMount}
          options={{
            fontSize: 13,
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            wordWrap: 'on',
            automaticLayout: true,
            padding: { top: 8 },
          }}
        />
      </div>
    </div>
  );
};

export default SqlEditor;
