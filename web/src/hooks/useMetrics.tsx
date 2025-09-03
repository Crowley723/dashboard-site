import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import type { ResultData } from '@/types/Data.ts';
import { fetchMetrics } from '@/api/Data.tsx';
import { processResult } from '@/utils/Data.tsx';

export const useAllMetrics = (
  options?: Omit<UseQueryOptions<ResultData[]>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: ['metrics', 'all'],
    queryFn: () => fetchMetrics(),
    staleTime: 5 * 60 * 1000, // 5 minutes
    refetchInterval: 30 * 1000, // Refetch every 30 seconds
    ...options,
  });
};

export const useMetricsQuery = (queries?: string[]) => {
  return useQuery({
    queryKey: ['metrics', queries],
    queryFn: () => fetchMetrics(queries),
    select: (data: ResultData[]) => {
      return data.map((result) => processResult(result)).filter(Boolean);
    },
  });
};

export const useMetrics = (
  queries: string[],
  options?: Omit<UseQueryOptions<ResultData[]>, 'queryKey' | 'queryFn'>
) => {
  return useQuery({
    queryKey: ['metrics', 'specific', ...queries.sort()],
    queryFn: () => fetchMetrics(queries),
    enabled: queries.length > 0,
    staleTime: 5 * 60 * 1000,
    refetchInterval: 30 * 1000,
    ...options,
  });
};

export const useMetric = (
  queryName: string,
  options?: Omit<
    UseQueryOptions<ResultData | undefined>,
    'queryKey' | 'queryFn'
  >
) => {
  return useQuery({
    queryKey: ['metrics', 'single', queryName],
    queryFn: async () => {
      const data = await fetchMetrics([queryName]);
      return data.find((item) => item.query_name === queryName);
    },
    enabled: !!queryName,
    staleTime: 5 * 60 * 1000,
    refetchInterval: 30 * 1000,
    ...options,
  });
};
