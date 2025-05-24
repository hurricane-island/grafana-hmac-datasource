import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData> {}

export function ConfigEditor(props: Props) {
  const { onOptionsChange, options } = props;
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        path: event.target.value,
      },
    });
  };

  // Secure field (only sent to the backend)
  const onAPIKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        secretKey: event.target.value,
      },
    });
  };

  const onResetAPIKey = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        secretKey: '',
      },
    });
  };

  // Secure field (only sent to the backend)
  const onClientIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      secureJsonData: {
        secretKey: event.target.value,
      },
    });
  };

  const onResetClientId = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        clientId: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        clientId: '',
      },
    });
  };

  return (
    <>
      <InlineField label="Path" labelWidth={14} interactive tooltip={'Json field returned to frontend'}>
        <Input
          id="config-editor-path"
          onChange={onPathChange}
          value={jsonData.path}
          placeholder="Enter the path, e.g. /api/v1"
          width={40}
        />
      </InlineField>
      <InlineField label="Client ID" labelWidth={14} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-client-id"
          isConfigured={secureJsonFields.clientId}
          value={secureJsonData?.clientId}
          placeholder="Enter your API key"
          width={40}
          onReset={onResetClientId}
          onChange={onClientIdChange}
        />
      </InlineField>
      <InlineField label="Secret Key" labelWidth={14} interactive tooltip={'Secure json field (backend only)'}>
        <SecretInput
          required
          id="config-editor-api-key"
          isConfigured={secureJsonFields.secretKey}
          value={secureJsonData?.secretKey}
          placeholder="Enter your API key"
          width={40}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
    </>
  );
}
