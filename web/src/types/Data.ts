export interface Annotations {
  node: string;
  __name__: string;
  app: string;
  cluster: string;
  component: string;
  container: string;
  job: string;
  k8s_cluster_name: string;
  namespace: string;
  pod: string;
  source: string;
  status: boolean;
  workload?: string;
}
export interface ValScalar {
  timestamp: number;
  metric: Annotations;
  value: number | string;
}

export interface ValVector {
  timestamp: number;
  metric: Annotations;
  value: [number, number];
}

export interface ValMatrix {
  timestamp: number;
  metrics: Annotations;
  value: [number, number][];
}

export type PrometheusData = ValScalar | ValVector[] | ValMatrix[];

export interface ResultData {
  query_name: string;
  type: 'scalar' | 'vector' | 'matrix';
  data: PrometheusData;
  timestamp: number;
}
