import React from 'react';
import Editor from '@monaco-editor/react';
import { Button } from '@/components/ui/button';
import { Play, Square } from 'lucide-react';
import { useQueryStore } from '@/stores/query';
import { SqlHintContext, useSqlHintStore } from '@/stores/sql-hints';
import { quoteIdent, resolveDialect } from '@/lib/sql-ident';
import { toast } from '@/components/ui/use-toast';

interface SqlEditorProps {
  connID: number;
  connectionType?: string;
  database?: string;
  schema?: string;
  contextLabel?: string;
  sql: string;
  onSqlChange: (sql: string) => void;
}

const SQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'IN', 'LIKE', 'IS', 'NULL',
  'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE',
  'CREATE', 'TABLE', 'DROP', 'ALTER', 'INDEX', 'VIEW', 'DATABASE',
  'JOIN', 'INNER', 'LEFT', 'RIGHT', 'OUTER', 'ON', 'AS',
  'GROUP', 'BY', 'ORDER', 'HAVING', 'LIMIT', 'OFFSET', 'DISTINCT',
  'UNION', 'ALL', 'CASE', 'WHEN', 'THEN', 'ELSE', 'END',
  'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'DEFAULT', 'CONSTRAINT',
  'IF', 'EXISTS', 'BETWEEN',
];

let completionProviderRegistered = false;
let latestHintContext: SqlHintContext = { connID: 0 };

/** 按方言给补全插入的标识符加引号（简单名不加）。 */
function quoteIdentifier(name: string, connectionType?: string): string {
  return quoteIdent(name, resolveDialect(connectionType));
}

function completionRange(model: any, position: any) {
  const word = model.getWordUntilPosition(position);
  return {
    startLineNumber: position.lineNumber,
    endLineNumber: position.lineNumber,
    startColumn: word.startColumn,
    endColumn: word.endColumn,
  };
}

function detectMemberAccess(model: any, position: any): string | null {
  const line = model.getLineContent(position.lineNumber).slice(0, position.column - 1);
  const match = line.match(/(?:`([^`]+)`|"([^"]+)"|([A-Za-z_][A-Za-z0-9_]*))\.\s*([A-Za-z_][A-Za-z0-9_]*)?$/);
  return match ? (match[1] || match[2] || match[3] || null) : null;
}

function registerSqlCompletionProvider(monaco: any) {
  if (completionProviderRegistered) return;
  completionProviderRegistered = true;

  monaco.languages.registerCompletionItemProvider('sql', {
    triggerCharacters: ['.', '`', '"'],
    provideCompletionItems: async (model: any, position: any) => {
      const range = completionRange(model, position);
      const ctx = latestHintContext;
      const store = useSqlHintStore.getState();
      const memberTable = detectMemberAccess(model, position);

      if (memberTable && ctx.connID && ctx.database) {
        const columns = await store.ensureTableColumns(ctx, memberTable);
        return {
          suggestions: columns.map((column, index) => ({
            label: column.name,
            kind: monaco.languages.CompletionItemKind.Field,
            insertText: quoteIdentifier(column.name, ctx.connectionType),
            range,
            detail: column.databaseType,
            documentation: [
              column.isPrimary ? 'Primary Key' : '',
              column.nullable ? 'Nullable' : 'Not Null',
              column.comment || '',
            ].filter(Boolean).join('\n'),
            sortText: `1_${String(index).padStart(4, '0')}_${column.name}`,
          })),
        };
      }

      const keywordSuggestions = SQL_KEYWORDS.map((keyword, index) => ({
        label: keyword,
        kind: monaco.languages.CompletionItemKind.Keyword,
        insertText: keyword,
        range,
        detail: 'SQL keyword',
        sortText: `3_${String(index).padStart(4, '0')}_${keyword}`,
      }));

      if (!ctx.connID || !ctx.database) {
        return { suggestions: keywordSuggestions };
      }

      const catalog = await store.ensureCatalog(ctx);
      if (!catalog) {
        return { suggestions: keywordSuggestions };
      }

      const tableSuggestions = catalog.tables.map((table, index) => ({
        label: table.name,
        kind: monaco.languages.CompletionItemKind.Struct,
        insertText: quoteIdentifier(table.name, ctx.connectionType),
        range,
        detail: 'table',
        documentation: table.comment || '',
        sortText: `1_${String(index).padStart(4, '0')}_${table.name}`,
      }));

      const viewSuggestions = catalog.views.map((view, index) => ({
        label: view.name,
        kind: monaco.languages.CompletionItemKind.Module,
        insertText: quoteIdentifier(view.name, ctx.connectionType),
        range,
        detail: 'view',
        documentation: view.definition || '',
        sortText: `2_${String(index).padStart(4, '0')}_${view.name}`,
      }));

      return {
        suggestions: [
          ...tableSuggestions,
          ...viewSuggestions,
          ...keywordSuggestions,
        ],
      };
    },
  });
}

/** 基础 SQL 格式化：分号后换行、关键字大写、去除多余空行 */
function formatSql(sql: string): string {
  const KEYWORDS = [
    'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'IN', 'LIKE', 'IS', 'NULL',
    'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET', 'DELETE',
    'CREATE', 'TABLE', 'DROP', 'ALTER', 'INDEX', 'VIEW', 'DATABASE',
    'JOIN', 'INNER', 'LEFT', 'RIGHT', 'OUTER', 'ON', 'AS',
    'GROUP', 'BY', 'ORDER', 'HAVING', 'LIMIT', 'OFFSET', 'DISTINCT',
    'UNION', 'ALL', 'CASE', 'WHEN', 'THEN', 'ELSE', 'END',
    'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'DEFAULT', 'CONSTRAINT',
    'IF', 'EXISTS', 'BETWEEN',
  ];
  let result = sql;
  // 关键字大写（仅边界单词）
  KEYWORDS.forEach(kw => {
    const re = new RegExp(`\\b${kw}\\b`, 'gi');
    result = result.replace(re, kw);
  });
  // 分号后换行
  result = result.replace(/;\s*/g, ';\n');
  // 去除连续多余空行
  result = result.replace(/\n{3,}/g, '\n\n');
  return result.trim() + (result.trim().endsWith(';') ? '' : '');
}

const SqlEditor: React.FC<SqlEditorProps> = ({ connID, connectionType, database, schema, contextLabel, sql, onSqlChange }) => {
  const { execute, executing } = useQueryStore();
  const editorRef = React.useRef<any>(null);
  const contextRef = React.useRef({ connID, database, schema });

  // 把当前激活 tab 的上下文同步给（全局单例的）补全 provider。
  React.useEffect(() => {
    latestHintContext = { connID, connectionType, database, schema };
  }, [connID, connectionType, database, schema]);

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
      maxRows: 500,
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
    // 注册 SQL 自动补全（幂等：completionProviderRegistered 防重复）
    registerSqlCompletionProvider(monaco);

    // Cmd/Ctrl + Enter 执行
    editor.addCommand(
      monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter,
      () => handleExecuteRef.current()
    );

    // 运行选中：右键菜单
    editor.addAction({
      id: 'cyancat.runSelection',
      label: '运行选中 SQL',
      contextMenuGroupId: 'cyancat',
      contextMenuOrder: 1,
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter],
      run: () => handleExecuteRef.current(),
    });

    // 格式化 SQL（基础版：将分号后换行，关键字大写）
    editor.addAction({
      id: 'cyancat.formatSql',
      label: '格式化 SQL',
      contextMenuGroupId: 'cyancat',
      contextMenuOrder: 2,
      keybindings: [monaco.KeyMod.Shift | monaco.KeyMod.Alt | monaco.KeyCode.KeyF],
      run: (ed: any) => {
        const model = ed.getModel();
        if (!model) return;
        const original = model.getValue() as string;
        const formatted = formatSql(original);
        ed.executeEdits('cyancat-format', [{
          range: model.getFullModelRange(),
          text: formatted,
        }]);
      },
    });

    // 注释/取消注释选中行
    editor.addAction({
      id: 'cyancat.toggleComment',
      label: '注释/取消注释',
      contextMenuGroupId: 'cyancat',
      contextMenuOrder: 3,
      keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.Slash],
      run: (ed: any) => {
        ed.getAction('editor.action.commentLine')?.run();
      },
    });
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
        <span className="ml-auto max-w-[45%] truncate text-xs text-muted-foreground" title={contextLabel || '未绑定执行上下文'}>
          {contextLabel || '未绑定执行上下文'}
        </span>
      </div>

      {/* 编辑器 */}
      <div className="flex-1">
        <Editor
          height="100%"
          language="sql"
          theme="vs-dark"
          value={sql}
          onChange={v => onSqlChange(v || '')}
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
