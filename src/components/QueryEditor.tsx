import React, { useState } from 'react';
import { Field, Stack, Combobox, ComboboxOption, MultiCombobox } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, ObservationQuery, ThingWithDataStreams, DataStream } from '../types';

// Data stream lookup by thing ID.
type DataStreams = Record<string, ComboboxOption[]>;

/**
 * Query uses backend data to populate interface with available
 * resource labels and identifiers.
 */
export function QueryEditor({
  // Current state of the query
  query,
  // Frontend data source instance
  datasource,
  // Query change handler
  onChange,
}: QueryEditorProps<DataSource, ObservationQuery, MyDataSourceOptions>) {
  // Options for multi-select to add to query
  const [dataStreamOptions, setDataStreamOptions] = useState<ComboboxOption[]>([]);
  // Topological mapping for rendering conditional options
  const [dataStreams, setDataStreams] = useState<DataStreams>({});
  // When parent ID changes, update options for child multi-select
  const onComboboxChange = (option: ComboboxOption) => {
    setDataStreamOptions(dataStreams[option.value]);
    onChange({ ...query, thingId: option.value });
  };
  // Format selection as query string for backend request
  const onMultiComboboxChange = (value: ComboboxOption[]) => {
    const queryString = value.map((each) => each.value).join(',');
    onChange({ ...query, dataStreamIds: queryString });
  };
  /**
   * Get and parse nodes to collect data using the datasource
   * resource API. This function is passed direct to the Combobox
   * component instead of a static list of options.
   */
  const options = (): Promise<ComboboxOption[]> =>
    datasource.getResource('sites').then(
      (resources: ThingWithDataStreams[]) => {
        const dataStreams: DataStreams = {};
        const selectThings = resources.map((each) => {
          const key = each.thing.id;
          dataStreams[key] = each.dataStreams.map((ds: DataStream) => {
            return {
              label: ds.name,
              value: ds.id,
            }});
          return {
            label: each.thing.name,
            value: key,
          };
        });
        setDataStreams(dataStreams);
        return selectThings;
      },
      (err) => {
        console.error({ err });
        return []
      }
    );
  return (
    <div>
      <Stack gap={0}>
        <Field label="Thing by ID">
          <Combobox id="query-editor-thing-id" options={options} onChange={onComboboxChange} />
        </Field>
        <Field label="Data Stream by ID">
          <MultiCombobox
            id="query-editor-data-stream-id"
            options={dataStreamOptions}
            onChange={onMultiComboboxChange}
            enableAllOption={true} // Allow selecting all data streams
          />
        </Field>
      </Stack>
    </div>
  );
}
