import { useMetricsQuery } from '@/hooks/useMetrics.tsx';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  ResponsiveContainer,
} from 'recharts';

export function ClusterAvgCPUCard() {
  const {
    data: metrics,
    isLoading,
    error,
    isError,
  } = useMetricsQuery(['total_cluster_cpu_perc']);

  if (isLoading) return <div>Loading CPU metrics...</div>;
  if (isError) return <div>Error loading metrics: {error.message}</div>;

  // Get the first matrix data (your CPU data)
  const matrixResult = metrics?.find((m) => m?.type === 'matrix');
  const cpuData = matrixResult?.processed?.[0]?.value || [];

  return (
    <div className="p-4">
      <h4 className="mb-4 text-sm font-medium">
        CPU Usage ({cpuData.length} points)
      </h4>

      <ResponsiveContainer width="100%" height={300}>
        <LineChart data={cpuData}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis
            dataKey={0} // Use index 0 for timestamp
            type="number"
            scale="time"
            domain={['dataMin', 'dataMax']}
            tickFormatter={(timestamp) =>
              new Date(timestamp * 1000).toLocaleTimeString()
            }
          />
          <YAxis
            dataKey={1} // Use index 1 for CPU value
            domain={['dataMin', 'dataMax']}
            tickFormatter={(value) => `${value.toFixed(1)}%`}
          />
          {/*<Tooltip*/}
          {/*  labelFormatter={(timestamp) =>*/}
          {/*    new Date(timestamp * 1000).toLocaleString()*/}
          {/*  }*/}
          {/*  formatter={(value) => [*/}
          {/*    `${parseFloat(value).toFixed(2)}%`,*/}
          {/*    'CPU Usage',*/}
          {/*  ]}*/}
          {/*/>*/}
          <Line
            type="monotone"
            dataKey={1} // Use index 1 for the CPU percentage values
            stroke="#8884d8"
            strokeWidth={2}
            dot={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
