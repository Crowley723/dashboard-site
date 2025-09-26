import { LineChart, Line, XAxis, YAxis, CartesianGrid } from 'recharts';
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
} from '@/components/ui/chart.tsx';

export interface ChartDataPoint {
  timestamp: number;
  [key: string]: number;
}

export interface ChartCardProps {
  title: string;
  data: ChartDataPoint[];
  dataKey: string;
  isLoading?: boolean;
  isError?: boolean;
  error?: Error;
  loadingText?: string;
  unit?: string; // e.g., '%', 'MB', 'GB', etc.
  color?: string; // CSS variable or color value
  strokeWidth?: number;
  showDots?: boolean;
  tickCount?: number;
  className?: string;
  chartClassName?: string;
}

const generateTicks = (min: number, max: number, count: number = 5) => {
  const range = max - min;
  if (!isFinite(range) || range <= 0 || count <= 1) {
    return [Number(min), Number(max)];
  }
  const step = range / (count - 1);
  const ticks: number[] = [];

  for (let i = 0; i < count; i++) {
    ticks.push(Number((min + step * i).toFixed(2)));
  }
  return ticks;
};

export function ChartCard({
  title,
  data,
  dataKey,
  isLoading = false,
  isError = false,
  error,
  loadingText = `Loading ${title.toLowerCase()}...`,
  unit = '%',
  color = 'var(--chart-1)',
  strokeWidth = 2,
  showDots = false,
  tickCount = 5,
  className = '',
  chartClassName = 'pr-[30px]',
}: ChartCardProps) {
  if (isLoading) return <div>Loading {loadingText}...</div>;
  if (isError)
    return (
      <div>
        Error loading {title.toLowerCase()}: {error?.message}
      </div>
    );

  if (!data || data.length === 0) {
    return <div>No data available for {title.toLowerCase()}</div>;
  }

  // Calculate dynamic domain
  const values = data.map((item) => item[dataKey]);
  const dataMax = Math.max(...values);
  const dataMin = Math.min(...values);
  const padding = (dataMax - dataMin) * 0.1;
  const yAxisMax = dataMax + padding;
  const yAxisMin = Math.max(0, dataMin - padding);

  const ticks = generateTicks(yAxisMin, yAxisMax, tickCount);

  // Create chart config dynamically
  const chartConfig = {
    [dataKey]: {
      label: title,
      color: color,
    },
  } satisfies ChartConfig;

  return (
    <div className={`flex-grow rounded-md border ${className}`}>
      <h4 className="p-4 text-sm font-medium">{title}</h4>

      <ChartContainer config={chartConfig} className={chartClassName}>
        <LineChart
          accessibilityLayer
          data={data}
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
            domain={[yAxisMin, yAxisMax]}
            ticks={ticks}
            tickLine={true}
            axisLine={true}
            tickMargin={8}
            tickFormatter={(value) => `${value.toFixed(1)}${unit}`}
          />
          <ChartTooltip
            content={({ active, payload, label }) => {
              if (!active || !payload || !payload.length) return null;

              const date = new Date(Number(label));
              const formattedTime = date.toLocaleString('en-US', {
                hour: '2-digit',
                minute: '2-digit',
              });

              const value = Number(payload[0].value).toFixed(1);

              return (
                <div className="rounded-lg border bg-background p-2 shadow-md w-64">
                  <div className="grid grid-cols-2">
                    <div className="flex flex-col min-w-0">
                      <span className="text-[0.70rem] uppercase text-muted-foreground break-words leading-tight">
                        Time
                      </span>
                      <span className="font-bold break-words text-sm">
                        {formattedTime}
                      </span>
                    </div>
                    <div className="flex flex-col min-w-0">
                      <span className="text-[0.70rem] uppercase text-muted-foreground break-words leading-tight">
                        {title}
                      </span>
                      <span className="font-bold break-words text-sm">
                        {value}
                        {unit}
                      </span>
                    </div>
                  </div>
                </div>
              );
            }}
          />
          <Line
            dataKey={dataKey}
            type="monotone"
            stroke={color}
            strokeWidth={strokeWidth}
            dot={showDots}
          />
        </LineChart>
      </ChartContainer>
    </div>
  );
}
