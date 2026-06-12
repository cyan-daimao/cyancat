import React from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Button } from '@/components/ui/button';
import { Download, Copy, FileText, Clipboard, Hash } from 'lucide-react';
import type { QueryResultDTO } from '@/lib/api/types';
import { toast } from '@/components/ui/use-toast';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { useSchemaStore } from '@/stores/schema';

interface ResultPanelProps {
  result: QueryResultDTO;
  /** 可选：用于生成 INSERT SQL 时的表名；不传则尝试从全局 schema store 取 */
  tableName?: string;
}

/** 将任意值格式化为字符串（NULL/undefined 统一显示为 "NULL"） */
const formatValue = (v: any): string => {
  if (v === null || v === undefined) return 'NULL';
  return String(v);
};

/** 把值序列化成 SQL 字面量（NULL/数字/布尔保持裸值，其余加单引号并转义） */
const toSqlLiteral = (v: any): string => {
  if (v === null || v === undefined) return 'NULL';
  if (typeof v === 'number' || typeof v === 'bigint') return String(v);
  if (typeof v === 'boolean') return v ? 'TRUE' : 'FALSE';
  const s = String(v).replace(/'/g, "''");
  return `'${s}'`;
};

const ResultPanel: React.FC<ResultPanelProps> = ({ result, tableName }) => {
  const parentRef = React.useRef<HTMLDivElement>(null);

  const cols = result.columns ?? [];
  const rows = result.rows ?? [];

  // 从全局 store 推断表名（INSERT SQL 用），若未传 prop
  const selectedNode = useSchemaStore(s => s.selectedNode);
  const resolvedTableName =
    tableName ||
    selectedNode?.tableName ||
    'table_name';

  const virtualizer = useVirtualizer({
    count: rows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 28,
    overscan: 20,
  });

  const exportCSV = () => {
    if (!rows.length) return;
    const header = cols.map(c => c.name).join(',');
    const body = rows.map(r => r.map(v => `"${String(v ?? '').replace(/"/g, '""')}"`).join(',')).join('\n');
    const csv = header + '\n' + body;
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = 'query_result.csv'; a.click();
    URL.revokeObjectURL(url);
    toast({ title: 'CSV 已导出' });
  };

  const copyAsMarkdown = () => {
    if (!rows.length) return;
    const header = '| ' + cols.map(c => c.name).join(' | ') + ' |';
    const sep = '| ' + cols.map(() => '---').join(' | ') + ' |';
    const body = rows.map(r => '| ' + r.map(v => String(v ?? 'NULL')).join(' | ') + ' |').join('\n');
    navigator.clipboard.writeText(header + '\n' + sep + '\n' + body);
    toast({ title: '已复制为 Markdown 表格' });
  };

  // ----- 右键菜单：复制操作 -----
  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast({ title: '已复制' });
    } catch (e: any) {
      toast({ title: '复制失败', description: e?.message, variant: 'destructive' });
    }
  };

  const copyCell = (rowIdx: number, colIdx: number) => {
    const v = rows[rowIdx]?.[colIdx];
    copyText(formatValue(v));
  };

  const copyRowTSV = (rowIdx: number) => {
    const row = rows[rowIdx];
    if (!row) return;
    copyText(row.map(formatValue).join('\t'));
  };

  const copyRowInsert = (rowIdx: number) => {
    const row = rows[rowIdx];
    if (!row) return;
    const colList = cols.map(c => `\`${c.name}\``).join(', ');
    const valList = row.map(toSqlLiteral).join(', ');
    const sql = `INSERT INTO \`${resolvedTableName}\` (${colList}) VALUES (${valList});`;
    copyText(sql);
  };

  const copyColName = (colIdx: number) => {
    const c = cols[colIdx];
    if (!c) return;
    copyText(c.name);
  };

  const renderCell = (value: any) => {
    if (value === null || value === undefined) {
      return <span className="text-muted-foreground italic">NULL</span>;
    }
    const str = String(value);
    if (str.startsWith('{') || str.startsWith('[')) {
      try {
        JSON.parse(str);
        return <span className="text-blue-400 text-xs font-mono">{str.length > 200 ? str.slice(0, 200) + '...' : str}</span>;
      } catch {}
    }
    return <span className="truncate">{str}</span>;
  };

  return (
    <div className="flex flex-col h-full">
      {/* 工具栏 */}
      <div className="flex items-center gap-2 px-2 py-1 border-b border-border shrink-0 text-xs text-muted-foreground">
        <span>{rows.length} 行</span>
        <span>·</span>
        <span>{cols.length} 列</span>
        {(result.durationMs ?? 0) > 0 && <><span>·</span><span>{result.durationMs}ms</span></>}
        {result.truncated && <><span>·</span><span className="text-amber-500">结果已截断</span></>}
        <div className="flex-1" />
        <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={exportCSV}>
          <Download className="h-3 w-3 mr-1" />CSV
        </Button>
        <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={copyAsMarkdown}>
          <Copy className="h-3 w-3 mr-1" />Markdown
        </Button>
      </div>

      {/* 表格 */}
      <div ref={parentRef} className="flex-1 overflow-auto">
        <table className="w-full text-xs">
          <thead className="sticky top-0 bg-muted/80 backdrop-blur-sm z-10">
            <tr>
              <th className="px-2 py-1 text-left font-medium text-muted-foreground w-10">#</th>
              {cols.map((col, i) => (
                <ContextMenu key={i}>
                  <ContextMenuTrigger asChild>
                    <th className="px-2 py-1 text-left font-medium text-muted-foreground whitespace-nowrap cursor-default">
                      {col.name}
                      <span className="ml-1 opacity-50">{col.databaseType}</span>
                    </th>
                  </ContextMenuTrigger>
                  <ContextMenuContent>
                    <ContextMenuItem onClick={() => copyColName(i)}>
                      <Hash className="h-3 w-3" />
                      复制列名
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>
              ))}
            </tr>
          </thead>
          <tbody>
            {virtualizer.getVirtualItems().map(virtualRow => {
              const rowIdx = virtualRow.index;
              const row = rows[rowIdx];
              return (
                <ContextMenu key={rowIdx}>
                  <ContextMenuTrigger asChild>
                    <tr
                      className="border-b border-border/50 hover:bg-accent/30"
                      style={{ height: virtualRow.size }}
                    >
                      <td className="px-2 py-0.5 text-muted-foreground">{rowIdx + 1}</td>
                      {row?.map((cell, ci) => (
                        <ContextMenu key={ci}>
                          <ContextMenuTrigger asChild>
                            <td className="px-2 py-0.5 max-w-xs">{renderCell(cell)}</td>
                          </ContextMenuTrigger>
                          <ContextMenuContent>
                            <ContextMenuItem onClick={() => copyCell(rowIdx, ci)}>
                              <Copy className="h-3 w-3" />
                              复制单元格
                            </ContextMenuItem>
                            <ContextMenuSeparator />
                            <ContextMenuItem onClick={() => copyRowTSV(rowIdx)}>
                              <Clipboard className="h-3 w-3" />
                              复制行 (TSV)
                            </ContextMenuItem>
                            <ContextMenuItem onClick={() => copyRowInsert(rowIdx)}>
                              <FileText className="h-3 w-3" />
                              复制行 (INSERT SQL)
                            </ContextMenuItem>
                          </ContextMenuContent>
                        </ContextMenu>
                      ))}
                    </tr>
                  </ContextMenuTrigger>
                  <ContextMenuContent>
                    <ContextMenuItem onClick={() => copyRowTSV(rowIdx)}>
                      <Clipboard className="h-3 w-3" />
                      复制行 (TSV)
                    </ContextMenuItem>
                    <ContextMenuItem onClick={() => copyRowInsert(rowIdx)}>
                      <FileText className="h-3 w-3" />
                      复制行 (INSERT SQL)
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default ResultPanel;
