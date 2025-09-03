import { createFileRoute } from '@tanstack/react-router';
import { useMetricsQuery } from '@/hooks/useMetrics.tsx';

export const Route = createFileRoute('/')({
  component: Index,
});

const SpecificMetricsComponent = () => {
  const { data: metrics, isLoading, error, isError } = useMetricsQuery(['up']);

  if (isLoading) return <div>Loading specific metrics...</div>;
  if (isError) return <div>Error loading metrics: {error.message}</div>;
  if (metrics) {
    console.log(metrics);
  }

  return (
    <div>
      <h2>Metrics Data</h2>
      {metrics?.map((processedResult, index) => (
        <div key={index}>
          <h3>Type: {processedResult?.type}</h3>

          {processedResult?.type === 'scalar' && (
            <div>
              <p>Value: {processedResult.processed.value}</p>
              <p>Timestamp: {processedResult.processed.timestamp}</p>
            </div>
          )}

          {processedResult?.type === 'vector' && (
            <div>
              {processedResult.processed.map((item, idx) => (
                <div key={idx}>
                  <p>Metric: {item.metric.app + '-' + item.metric.component}</p>
                  <p>Value: {item.value.value ? 'Up' : 'Down'}</p>
                </div>
              ))}
            </div>
          )}

          {processedResult?.type === 'matrix' && (
            <div>
              {processedResult.processed.map((series, seriesIdx) => (
                <div key={seriesIdx}>
                  <h4>Series: {JSON.stringify(series.metric)}</h4>
                  {series.value.map((point, pointIdx) => (
                    <p key={pointIdx}>
                      {point.value.timestamp}: {point.value.value}
                    </p>
                  ))}
                </div>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

function Index() {
  return (
    <div className="p-2">
      <h3>Welcome Home!</h3>
      <SpecificMetricsComponent />
    </div>
  );
}
