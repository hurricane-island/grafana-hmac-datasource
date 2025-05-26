import { test, expect } from '@grafana/plugin-e2e';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

test('smoke: should render config editor', async ({ 
  createDataSourceConfigPage, 
  readProvisionedDataSource, 
  page
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByLabel('Base Path')).toBeVisible();
});

test('"Save & test" should be successful when configuration is valid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<MyDataSourceOptions, MySecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Server URL' }).fill(ds.jsonData.basePath ?? '');
  await page.getByRole('textbox', { name: 'Base Path' }).fill(ds.jsonData.basePath ?? '');
  await page.getByRole('textbox', { name: 'Auth Method' }).fill(ds.jsonData.authMethod ?? '');
  await page.getByRole('textbox', { name: 'Client ID' }).fill(ds.secureJsonData?.secretKey ?? '');
  await page.getByRole('textbox', { name: 'Secret Key' }).fill(ds.secureJsonData?.secretKey ?? '');
  await expect(configPage.saveAndTest()).toBeOK();
});

test('"Save & test" should fail when configuration is invalid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<MyDataSourceOptions, MySecureJsonData>({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByRole('textbox', { name: 'Base Path' }).fill(ds.jsonData.basePath ?? '');
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', { hasText: 'Secret Key is missing' });
});
