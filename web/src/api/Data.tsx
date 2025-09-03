import type { ResultData } from '@/types/Data.ts';

export const fetchMetrics = async (
  queries?: string[]
): Promise<ResultData[]> => {
  const url = new URL('/api/data', window.location.origin);

  if (queries && queries.length > 0) {
    url.searchParams.set('queries', queries.join(','));
  }

  const response = await fetch(url.toString(), {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch metrics: ${response.statusText}`);
  }

  return response.json();
};
