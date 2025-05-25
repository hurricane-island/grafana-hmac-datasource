import React, { ChangeEvent } from 'react';
import { InlineField, Input, Stack } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, ObservationQuery } from '../types';

type Props = QueryEditorProps<DataSource, ObservationQuery, MyDataSourceOptions>;

export function QueryEditor({ query, onChange }: Props) {
  const onDataStreamIdsChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, datastreamIds: event.target.value });
  };
  return (
    <Stack gap={0}>
      <InlineField label="Data Stream IDs">
        <Input
          id="query-editor-data-stream-ids"
          onChange={onDataStreamIdsChange}
          value={query.datastreamIds}
          placeholder='one,two'
          required
        />
      </InlineField>
    </Stack>
  );
}
