import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import { ChartCard } from '@/components/LineChartCard.tsx';

export function TraefikTotalRequestsHourCard() {
  const {
    data: metrics,
    isLoading,
    error,
    isError,
  } = useMetricsQuery(['traefik_requests_total_1h']);

  const matrixResult = metrics?.find((m) => m?.type === 'matrix');
  const rawData = matrixResult?.processed?.[0]?.values || [];

  const data = rawData.map(([timestamp, value]) => ({
    timestamp: Number(timestamp) * 1000,
    requests: value,
  }));

  return (
    <ChartCard
      title="Total Requests"
      data={data}
      dataKey="requests"
      isLoading={isLoading}
      isError={isError}
      error={error || undefined}
      unit=""
      valueDecimals={1}
      color="var(--chart-2)"
    />
  );
}
