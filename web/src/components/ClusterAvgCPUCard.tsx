import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import { LineChart, Line, XAxis, YAxis, CartesianGrid } from 'recharts';
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
} from '@/components/ui/chart.tsx';

const chartConfig = {
  cpu: {
    label: 'CPU Usage  ',
    color: 'var(--chart-1)',
  },
} satisfies ChartConfig;

export function ClusterAvgCPUCard() {
  const {
    data: metrics,
    isLoading,
    error,
    isError,
  } = useMetricsQuery(['total_cluster_cpu_perc']);

  if (isLoading) return <div>Loading CPU metrics...</div>;
  if (isError) return <div>Error loading metrics: {error.message}</div>;

  const matrixResult = metrics?.find((m) => m?.type === 'matrix');
  const rawData = matrixResult?.processed?.[0]?.values || [];

  const cpuData = rawData.map(([timestamp, value]) => ({
    timestamp: Number(timestamp) * 1000,
    cpu: Number(value).toFixed(2),
  }));

  return (
    <div className={'flex-grow rounded-md border'}>
      <h4 className="p-4 text-sm font-medium">CPU Usage (%)</h4>

      <ChartContainer config={chartConfig} className={'  pr-[30px]'}>
        <LineChart
          accessibilityLayer
          data={cpuData}
          margin={{
            left: 12,
            right: 12,
          }}
        >
          <CartesianGrid vertical={false} />
          <XAxis
            dataKey="timestamp"
            tickLine={true}
            axisLine={true}
            tickMargin={8}
            interval="equidistantPreserveStart"
            tickFormatter={(timestamp) => {
              const date = new Date(Number(timestamp));
              return date.toLocaleString('en-US', {
                hour: '2-digit',
                minute: '2-digit',
              });
            }}
          />
          <YAxis
            tickLine={false}
            axisLine={false}
            tickMargin={8}
            tickFormatter={(value) => `${value.toFixed(1)}% `}
          />
          <ChartTooltip
            content={({ active, payload, label }) => {
              if (!active || !payload || !payload.length) return null;

              const date = new Date(Number(label));
              const formattedTime = date.toLocaleString('en-US', {
                hour: '2-digit',
                minute: '2-digit',
              });

              return (
                <div className="rounded-lg border bg-background p-2 shadow-md">
                  <div className="grid grid-cols-2 gap-2">
                    <div className="flex flex-col">
                      <span className="text-[0.70rem] uppercase text-muted-foreground">
                        Time
                      </span>
                      <span className="font-bold">{formattedTime}</span>
                    </div>
                    <div className="flex flex-col">
                      <span className="text-[0.70rem] uppercase text-muted-foreground">
                        CPU Usage
                      </span>
                      <span className="font-bold">{payload[0].value}%</span>
                    </div>
                  </div>
                </div>
              );
            }}
          />
          <Line
            dataKey={'cpu'}
            type="monotone"
            stroke={`var(--color-cpu)`}
            strokeWidth={2}
            dot={false}
          />
        </LineChart>
      </ChartContainer>
    </div>
  );
}
