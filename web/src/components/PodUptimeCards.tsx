import { ScrollArea } from '@/components/ui/scroll-area';
import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import { Card } from '@/components/ui/card.tsx';

export function PodUptimeCards() {
  const { data: metrics, isLoading, error, isError } = useMetricsQuery(['up']);

  if (isLoading) return <div>Loading specific metrics...</div>;
  if (isError) return <div>Error loading metrics: {error.message}</div>;

  const allPods =
    metrics?.flatMap((processedResult) => {
      if (processedResult?.type === 'vector') {
        return processedResult.processed.map((item) => ({
          name: `${item.metric.app}-${item.metric.component}`,
          status: item.value[1] ? 'Up' : 'Down',
          timestamp: item.value[0],
          app: item.metric.app,
          component: item.metric.component,
        }));
      }
      return [];
    }) || [];

  const groupedPods = allPods.reduce(
    (acc, pod) => {
      const key = pod.name;
      if (!acc[key]) {
        acc[key] = {
          name: key,
          app: pod.app,
          component: pod.component,
          upCount: 0,
          totalCount: 0,
          pods: [],
        };
      }
      acc[key].totalCount++;
      if (pod.status === 'Up') {
        acc[key].upCount++;
      }
      acc[key].pods.push(pod);
      return acc;
    },
    {} as Record<
      string,
      {
        name: string;
        app: string;
        component: string;
        upCount: number;
        totalCount: number;
        pods: typeof allPods;
      }
    >
  );

  const podGroups = Object.values(groupedPods).sort((a, b) => {
    const aIsFullyUp = a.upCount === a.totalCount;
    const bIsFullyUp = b.upCount === b.totalCount;

    if (!aIsFullyUp && bIsFullyUp) return -1;
    if (aIsFullyUp && !bIsFullyUp) return 1;

    return a.app.localeCompare(b.app);
  });

  return (
    <ScrollArea className="h-152 w-96 rounded-md border">
      <div className="p-4">
        <h4 className="mb-4 text-sm leading-none font-medium">
          Pod Groups ({podGroups.length})
        </h4>
        <div className="space-y-2">
          {podGroups.map((group, index) => {
            const isFullyUp = group.upCount === group.totalCount;
            const isFullyDown = group.upCount === 0;
            const statusColor = isFullyUp
              ? 'bg-green-500'
              : isFullyDown
                ? 'bg-red-500'
                : 'bg-orange-500';
            const statusTextColor = isFullyUp
              ? 'text-green-600'
              : isFullyDown
                ? 'text-red-600'
                : 'text-orange-600';

            return (
              <Card key={`${group.name}-${index}`} className="p-3">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <span className="font-medium text-sm">{group.app}</span>
                    <span className="text-xs text-muted-foreground">
                      {group.component}
                    </span>
                  </div>
                  <div className="flex items-center space-x-2">
                    <div
                      className={`h-2 w-2 rounded-full ${statusColor}`}
                    ></div>
                    <span className={`text-xs font-medium ${statusTextColor}`}>
                      {group.upCount}/{group.totalCount}
                    </span>
                  </div>
                </div>
              </Card>
            );
          })}
          {podGroups.length === 0 && (
            <div className="text-center text-sm text-muted-foreground py-8">
              No pod groups found
            </div>
          )}
        </div>
      </div>
    </ScrollArea>
  );
}
