# How to use

1. First run

   ```bash
   npm i
   ```

2. Deploy Pachyderm Enterprise Edition. Community Edition will not work because the setup script uses over the maxmium number of pipelines.

3. Disable auth on the Pachyderm cluster if it is not already.

   ```bash
   pachctl auth deactivate
   ```

4. Then at the root of the project run

   ```bash
   npm run throughput:setup
   ```

## Run REST API throughput test

!IMPORTANT!

To have inspectJob and inspectJobSet pass, copy a job ID from pipeline 'dag-0-pipeline-0' and paste it into `const JOB_ID = ''` found in `/utils/rest.ts`

In the root of the project run

```bash
npm run throughput:test
```

## Run GraphQL API throughput test (deprecated)

(The build will fail unless you are on an older version that still has the graphql code)

In the root of the project run one of the following (depending on your env):

```bash
npm run throughput:test:graphql
```
