## Getting Started

```bash
npm install
```

## Setting up a performance cluster
- The managerCluster.ts script uses pachctl, so make sure your context is pointed to the cluster you want to operate on
- Run `npm run cluster:small:distributed` to see how things are ultimately set up
- To clean up the cluster, run `npm run cluster:teardown`
- Once you are confident, feel free to run `npm run cluster:medium:distributed` or `npm run cluster:large:distributed`
- If you want to set up a cluster of a custom size run `npm run cluster:setup -- ` with any of the following flags:

| Arg  | Type | Default |
| ------------- | ------------- | ------------- |
| distributed  | Boolean  | false |
| numProjects | Number | 10 |
| numRepos | Number | 10 |
| numPipelines | Number | 10 |
| numCommits | Number | 10 |
| numFiles | Number | 100 |
| sizeFileLarge | Number | 10000000 |
| sizeFileSmall | Number | 1000 |
| sizeLogsLarge | Number | 1000000 |
| sizePipelineWorkersSmall | Number | 1 |
| sizePipelineCPUSmall | Number | 0.01 |
| sizePipelineCPULarge | Number | 1 |
| sizePipelineMemorySmall | String | 100MB |
| sizePipelineMemoryLarge | String | 1GB |
| numInaccessibleProjects | Number | 0 |
