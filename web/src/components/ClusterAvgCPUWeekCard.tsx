import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import { ChartCard } from '@/components/LineChartCard.tsx';

export function ClusterAvgCPUWeekCard() {
  const {
    data: metrics,
    isLoading,
    error,
    isError,
  } = useMetricsQuery(['total_cluster_cpu_perc_7d']);

  const matrixResult = metrics?.find((m) => m?.type === 'matrix');
  const rawData = matrixResult?.processed?.[0]?.values || [];

  const cpuData = rawData.map(([timestamp, value]) => ({
    timestamp: Number(timestamp) * 1000,
    cpu: Number(value),
  }));

  return (
    <ChartCard
      title="CPU Usage (%) (1 week)"
      data={cpuData}
      dataKey="cpu"
      isLoading={isLoading}
      isError={isError}
      error={error || undefined}
      unit="%"
      color="var(--chart-5)"
    />
  );
}
