import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import { ScrollArea } from '@/components/ui/scroll-area.tsx';
import { Card } from '@/components/ui/card.tsx';
import { Box, Crown } from 'lucide-react';

export function NodeStatusCard() {
  const {
    data: metrics,
    isLoading,
    error,
    isError,
  } = useMetricsQuery(['node_status']);

  if (isLoading) return <div>Loading specific metrics...</div>;
  if (isError) return <div>Error loading metrics: {error.message}</div>;

  const allNodes =
    metrics?.flatMap((processedResult) => {
      if (processedResult?.type === 'vector') {
        return processedResult.processed.map((item) => ({
          name: `${item.metric.node}`,
          status: item.value[1] ? 'Up' : 'Down',
          timestamp: item.value[0],
          master: item.metric.node.includes('master'),
        }));
      }
      return [];
    }) || [];

  const sortedNodes = allNodes.sort((a, b) => {
    if (a.status === 'Down' && b.status !== 'Down') {
      return -1;
    }
    if (a.status !== 'Down' && b.status === 'Down') {
      return 1;
    }

    if (a.master && !b.master) return -1;
    if (!a.master && b.master) return 1;

    return a.name.localeCompare(b.name, undefined, {
      numeric: true,
    });
  });

  return (
    <ScrollArea className="h-72 flex-grow rounded-md border">
      <div className="p-4">
        <h4 className="mb-4 text-sm leading-none font-medium">
          Node Status ({sortedNodes.length})
        </h4>
        <div className="space-y-2">
          {sortedNodes.map((node, index) => {
            const statusColor =
              node.status === 'Up' ? 'bg-green-500' : 'bg-orange-500';

            return (
              <Card key={`${node.name}-${index}`} className="p-3">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <div className="flex items-center gap-2">
                      {node.master ? (
                        <Crown className="h-4 w-4" />
                      ) : (
                        <Box className="h-4 w-4" />
                      )}
                      <span className="font-medium text-sm">{node.name}</span>
                    </div>
                    <span className="text-xs text-muted-foreground"></span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div
                      className={`h-2 w-2 rounded-full ${statusColor}`}
                    ></div>
                  </div>
                </div>
              </Card>
            );
          })}
          {sortedNodes.length === 0 && (
            <div className="text-center text-sm text-muted-foreground py-8">
              No pods found
            </div>
          )}
        </div>
      </div>
    </ScrollArea>
  );
}
