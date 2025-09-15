import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import { ScrollArea } from '@/components/ui/scroll-area.tsx';
import { Card } from '@/components/ui/card.tsx';

export function PodsPerNamespaceCard() {
  const {
    data: metrics,
    isLoading,
    isError,
  } = useMetricsQuery(['pods_running_per_namespace']);

  const allNamespaces =
    metrics?.flatMap((processedResult) => {
      if (processedResult?.type === 'vector') {
        return processedResult.processed.map((item) => ({
          name: `${item.metric.namespace}`,
          count: Number(item.value[1]),
          timestamp: item.value[0],
        }));
      }
      return [];
    }) || [];

  if (isLoading) return <div>Loading...</div>;
  if (isError) return <div>Error loading card.</div>;

  const sortedNamespaces = allNamespaces.sort((a, b) => {
    if (a.count > b.count) {
      return -1;
    }
    if (a.count < b.count) {
      return 1;
    }

    return a.name.localeCompare(b.name);
  });

  return (
    <ScrollArea className="h-72 w-96 rounded-md border">
      <div className="p-4">
        <h4 className="mb-4 text-sm leading-none font-medium">
          Pods Per Namespace
        </h4>
        <div className="space-y-2">
          {sortedNamespaces.map((namespace, index) => {
            return (
              <Card key={`${namespace.name}-${index}`} className="p-3">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <span className="font-medium text-sm">
                      {namespace.name}
                    </span>
                    <span className="text-xs text-muted-foreground"></span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div className={`h-2 w-2 rounded-full`}></div>
                    <span className={`text-xs font-medium`}>
                      {namespace.count}
                    </span>
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      </div>
    </ScrollArea>
  );
}
