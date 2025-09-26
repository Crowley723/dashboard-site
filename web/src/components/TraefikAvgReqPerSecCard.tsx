import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import { ChartCard } from '@/components/LineChartCard.tsx';

export function TraefikAvgReqPerSecCard() {
  const {
    data: metrics,
    isLoading,
    error,
    isError,
  } = useMetricsQuery(['traefik_requests_5m_avg']);

  const matrixResult = metrics?.find((m) => m?.type === 'matrix');
  const rawData = matrixResult?.processed?.[0]?.values || [];

  const data = rawData.map(([timestamp, value]) => ({
    timestamp: Number(timestamp) * 1000,
    requests: value,
  }));

  return (
    <ChartCard
      title="Average Requests Per Second"
      data={data}
      dataKey="requests"
      isLoading={isLoading}
      isError={isError}
      error={error || undefined}
      unit=""
      color="var(--chart-1)"
    />
  );
}
