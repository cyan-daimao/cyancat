import React from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import SqlEditor from './SqlEditor';
import ResultPanel from '@/components/data-table/ResultPanel';
import { useQueryStore } from '@/stores/query';
import { useSchemaStore } from '@/stores/schema';

const SqlWorkspace: React.FC = () => {
  const { results, activeResultIndex, setActiveResult, closeResult } = useQueryStore();
  const selectedNode = useSchemaStore(s => s.selectedNode);

  return (
    <div className="flex flex-col h-full">
      {/* SQL 编辑器 */}
      <div className="h-64 shrink-0 border-b border-border">
        <SqlEditor
          connID={selectedNode?.connID || 0}
          database={selectedNode?.database}
          schema={selectedNode?.schema}
        />
      </div>

      {/* 结果集 Tabs */}
      <div className="flex-1 overflow-hidden">
        {results.length === 0 ? (
          <div className="flex items-center justify-center h-full text-sm text-muted-foreground">
            执行 SQL 查询查看结果
          </div>
        ) : (
          <Tabs value={String(activeResultIndex)} onValueChange={v => setActiveResult(Number(v))}>
            <TabsList className="h-7">
              {results.map((r, i) => (
                <TabsTrigger key={i} value={String(i)} className="text-xs h-5 px-2">
                  结果 {i + 1}
                  <span className="ml-1 text-muted-foreground">({r.rows?.length || 0} 行)</span>
                </TabsTrigger>
              ))}
            </TabsList>
            {results.map((r, i) => (
              <TabsContent key={i} value={String(i)} className="h-full">
                <ResultPanel result={r} />
              </TabsContent>
            ))}
          </Tabs>
        )}
      </div>
    </div>
  );
};

export default SqlWorkspace;
