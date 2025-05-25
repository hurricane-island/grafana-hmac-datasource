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
        basePath: event.target.value,
      },
    });
  };

  const onServerUrlChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        serverUrl: event.target.value,
      },
    });
  };

    const onAuthMethodChange = (event: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({
      ...options,
      jsonData: {
        ...jsonData,
        authMethod: event.target.value,
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
        secretKey: false,
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
        clientId: event.target.value,
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
      <InlineField label="Server URL" labelWidth={14} interactive tooltip={'URL of server to use'}>
        <Input
          id="config-editor-server-url"
          onChange={onServerUrlChange}
          value={jsonData.serverUrl}
          placeholder="Enter the server URL"
          width={40}
        />
      </InlineField>
      <InlineField label="Base Path" labelWidth={14} interactive tooltip={'base API route'}>
        <Input
          id="config-editor-base-path"
          onChange={onPathChange}
          value={jsonData.basePath}
          placeholder="Enter the base API path"
          width={40}
        />
      </InlineField>
      <InlineField label="Auth Method" labelWidth={14} interactive tooltip={'Name of receiving service'}>
        <Input
          id="config-editor-auth-method"
          onChange={onAuthMethodChange}
          value={jsonData.authMethod}
          placeholder="..."
          width={40}
        />
      </InlineField>
      <InlineField label="Client ID" labelWidth={14} interactive tooltip={'Service routing key'}>
        <SecretInput
          required
          id="config-editor-client-id"
          isConfigured={secureJsonFields.clientId}
          value={secureJsonData?.clientId}
          placeholder="..."
          width={40}
          onReset={onResetClientId}
          onChange={onClientIdChange}
        />
      </InlineField>
      <InlineField label="Secret Key" labelWidth={14} interactive tooltip={'HMAC signing key'}>
        <SecretInput
          required
          id="config-editor-api-key"
          isConfigured={secureJsonFields.secretKey}
          value={secureJsonData?.secretKey}
          placeholder="..."
          width={40}
          onReset={onResetAPIKey}
          onChange={onAPIKeyChange}
        />
      </InlineField>
    </>
  );
}
