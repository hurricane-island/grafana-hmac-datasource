# Grafana data source for Sensor Things API HMAC authentication

(c) 2025 Hurricane Island Center for Science and Leadership

## About

This repository contains the source code for a Grafana backend data source that will integrate with an API
that requires hash-based message authentication (HMAC). The project was scaffolded using Grafana plugin tools,
and derived from the [backend datasource example](https://github.com/grafana/grafana-plugin-examples/tree/master/examples/datasource-with-backend#readme).

Source code and documentation is provided under the Apache 2 open source license without guarantee or implied warrantee. 

## Backend

The backend is written in Go, and uses the [Grafana Backend SDK](https://grafana.com/developers/plugin-tools/key-concepts/backend-plugins/grafana-plugin-sdk-for-go). Go dependencies are managed in `go.mod`, and source code can be found in the `pkg` directory.

### Dependencies

You can update the SDK to the latest version with Go commands:

```bash
go get -u github.com/grafana/grafana-plugin-sdk-go
go mod tidy
```

### Build

Following the Grafana template, we use Mage to build the binaries. The build process is controlled by `Magefile.go`.

For local Docker-based development we generally need `linux:ARM64`. You can build for distribution with `mage -v` or for Docker-based development with `mage -v linux:ARM64`. Calling `mage -l` lists available build targets and other commands.

### Testing

Plugins have a `Save & Test` button in the Grafana UI. The behavior is described by `pkg/datasource_test.go`.

## Frontend

1. Install dependencies

   ```bash
   npm install
   ```

2. Build plugin in development mode and run in watch mode

   ```bash
   npm run dev
   ```

3. Build plugin in production mode

   ```bash
   npm run build
   ```

4. Run the tests (using Jest)

   ```bash
   # Runs the tests and watches for changes, requires git init first
   npm run test

   # Exits after running all the tests
   npm run test:ci
   ```

5. Spin up a Grafana instance and run the plugin inside it (using Docker)

   ```bash
   npm run server
   ```

6. Run the E2E tests (using Playwright)

   ```bash
   # Spins up a Grafana instance first that we tests against
   npm run server

   # If you wish to start a certain Grafana version. If not specified will use latest by default
   GRAFANA_VERSION=11.3.0 npm run server

   # Starts the tests
   npm run e2e
   ```

7. Run the linter

   ```bash
   npm run lint

   # or

   npm run lint:fix
   ```

## Distribution

The plugin must be signed so Grafana can verify its authenticity. This is done with the `@grafana/sign-plugin` package, which includes commands and workflows to distribute through the Grafana plugins catalog.

The terms and process are explained in [plugin publishing and signing criteria](https://grafana.com/legal/plugins/#plugin-publishing-and-signing-criteria), and the [plugin signature levels documentation](https://grafana.com/legal/plugins/#what-are-the-different-classifications-of-plugins).

1. Create a [Grafana Cloud account](https://grafana.com/signup).
2. Make sure that the first part of the plugin ID matches the slug of your Grafana Cloud account.
   - _You can find the plugin ID in the `plugin.json` file inside your plugin directory. For example, if your account slug is `acmecorp`, you need to prefix the plugin ID with `acmecorp-`._
3. Create a Grafana Cloud API key with the `PluginPublisher` role.
4. Keep a record of this API key as it will be required for signing a plugin

### Signing with Github actions

If the plugin is using the github actions supplied with `@grafana/create-plugin` signing a plugin is included out of the box. The [release workflow](./.github/workflows/release.yml) can prepare everything to make submitting your plugin to Grafana as easy as possible. Before being able to sign the plugin however a secret needs adding to the Github repository.

1. Please navigate to "settings > secrets > actions" within your repo to create secrets.
2. Click "New repository secret"
3. Name the secret "GRAFANA_API_KEY"
4. Paste your Grafana Cloud API key in the Secret field
5. Click "Add secret"

#### Push a version tag

To trigger the workflow we need to push a version tag to github. This can be achieved with the following steps:

1. Run `npm version <major|minor|patch>`
2. Run `git push origin main --follow-tags`

## Learn more

Below you can find source code for existing app plugins and other related documentation.

- [`plugin.json` documentation](https://grafana.com/developers/plugin-tools/reference/plugin-json)
- [How to sign a plugin?](https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin)

## Generative AI Disclosure

Some of the code in this repository was created or edited using generative AI products including Github Copilot and Warp Terminal.
