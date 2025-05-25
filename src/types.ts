import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface ObservationQuery extends DataQuery {
  queryText?: string;
  datastreamId: string;
}

export const DEFAULT_QUERY: Partial<ObservationQuery> = {
  datastreamId: "",
};

export interface Observation {
  phenomenonTime: number;
  value: number;
}

export interface DataSourceResponse {
  observations: Observation[];
}

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  path?: string;
  serverUrl?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  secretKey?: string;
  clientId?: string;
}
