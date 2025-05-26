import React from 'react';
import { InlineField, Stack, Combobox, ComboboxOption } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, ObservationQuery, Thing } from '../types';

type Props = QueryEditorProps<DataSource, ObservationQuery, MyDataSourceOptions>;

// Query uses backend data to populate interface with available
// resource labels and identifiers.
export function QueryEditor({ query, datasource, onChange }: Props) {
  const onComboboxChange = (option: ComboboxOption) => {
    onChange({ ...query, thingId: option.value });
  };
  // Get and parse available things to collect data streams from.
  // Each query will have a single thing, but can request multiple
  // data streams within a time range.
  const options = () => datasource.getResource("sites").then(
    (things: Thing[]) => {
      return things.map((thing) => {
        return {
          label: thing.name,
          value: thing.id
        } as ComboboxOption
      })
    }, 
    (err) => {
      console.error({err})
      return [] as ComboboxOption[];
    })
  return (
    <div>
      <Stack gap={0}>
        <InlineField label="Thing">
          <Combobox 
            id="query-editor-thing-id"
            options={options}
            onChange={onComboboxChange}
            loading={true}
          >
          </Combobox>
        </InlineField>
      </Stack>
    </div>
  );
}
