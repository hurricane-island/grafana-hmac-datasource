import React, { ChangeEvent } from 'react';
import { InlineField, Input, Stack } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, ObservationQuery } from '../types';

type Props = QueryEditorProps<DataSource, ObservationQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, queryText: event.target.value });
  };

  const onDatastreamChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, datastreamId: event.target.value });
    // executes the query
    onRunQuery();
  };

  const { queryText, datastreamId } = query;

  return (
    <Stack gap={0}>
      <InlineField label="Constant">
        <Input
          id="query-editor-constant"
          onChange={onDatastreamChange}
          value={datastreamId}
        />
      </InlineField>
      <InlineField label="Query Text" labelWidth={16} tooltip="Not used yet">
        <Input
          id="query-editor-query-text"
          onChange={onQueryTextChange}
          value={queryText || ''}
          required
          placeholder="Enter a query"
        />
      </InlineField>
    </Stack>
  );
}
