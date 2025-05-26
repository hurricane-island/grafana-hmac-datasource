import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export interface ObservationQuery extends DataQuery {
  thingId: string;
}

export type Thing = {
  id: string
  name: string
  description?: string
  location: Array<{
    latitude: number
    longitude: number
  }>
}


export const DEFAULT_QUERY: Partial<ObservationQuery> = {
  thingId: "",
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  basePath?: string
  serverUrl?: string
  authMethod?: string
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  secretKey?: string;
  clientId?: string;
}
