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
  valueDecimals?: number;
  color?: string; // CSS variable or color value
  strokeWidth?: number;
  showDots?: boolean;
  tickCount?: number;
  className?: string;
  chartClassName?: string;
  // Optional: override automatic time formatting
  timeFormat?: 'auto' | 'minutes' | 'hours' | 'days' | 'months';
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

const determineTimeFormat = (timestamps: number[]) => {
  if (timestamps.length < 2) return 'minutes';

  const minTime = Math.min(...timestamps);
  const maxTime = Math.max(...timestamps);
  const rangeMs = maxTime - minTime;

  // Convert to different time units
  //const rangeMinutes = rangeMs / (1000 * 60);
  const rangeHours = rangeMs / (1000 * 60 * 60);
  const rangeDays = rangeMs / (1000 * 60 * 60 * 24);
  const rangeMonths = rangeMs / (1000 * 60 * 60 * 24 * 30);

  if (rangeMonths > 2) return 'months';
  if (rangeDays > 7) return 'days';
  if (rangeHours > 6) return 'hours';
  return 'minutes';
};

const formatTimestamp = (timestamp: number, format: string) => {
  const date = new Date(Number(timestamp));

  switch (format) {
    case 'months':
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
      });
    case 'days':
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
      });
    case 'hours':
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
      });
    case 'minutes':
    default:
      return date.toLocaleString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
      });
  }
};

const formatTooltipTimestamp = (timestamp: number, format: string) => {
  const date = new Date(Number(timestamp));

  switch (format) {
    case 'months':
      return date.toLocaleString('en-US', {
        month: 'long',
        day: 'numeric',
        year: 'numeric',
      });
    case 'days':
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    case 'hours':
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    case 'minutes':
    default:
      return date.toLocaleString('en-US', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
  }
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
  valueDecimals = 2,
  color = 'var(--chart-1)',
  strokeWidth = 2,
  showDots = false,
  tickCount = 5,
  className = '',
  chartClassName = 'pr-[30px]',
  timeFormat = 'auto',
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

  // Determine time format
  const timestamps = data.map((item) => item.timestamp);
  const actualTimeFormat =
    timeFormat === 'auto' ? determineTimeFormat(timestamps) : timeFormat;

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
            interval={Math.ceil(data.length / 5) - 1}
            tickFormatter={(timestamp) =>
              formatTimestamp(timestamp, actualTimeFormat)
            }
          />
          <YAxis
            domain={[yAxisMin, yAxisMax]}
            ticks={ticks}
            tickLine={true}
            axisLine={true}
            tickMargin={8}
            tickFormatter={(value) => `${value.toFixed(valueDecimals)}${unit}`}
          />
          <ChartTooltip
            content={({ active, payload, label }) => {
              if (!active || !payload || !payload.length) return null;

              const formattedTime = formatTooltipTimestamp(
                Number(label),
                actualTimeFormat
              );
              const value = Number(payload[0].value).toFixed(valueDecimals);

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
