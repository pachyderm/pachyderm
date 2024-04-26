## Getting Started

```bash
npm install
```

## Running load-tests against a cluster
- Run the load script with `npm run load -- --browserCount=10 --baseURL=http://104.154.210.180/`
- If you want to run the load tests with custom values, run `npm run load -- ` with any of the following flags:

| Arg  | Type | Default |
| ------------- | ------------- | ------------- |
| baseURL  | string  | http://localhost:4000 |
| testRuntime | Number | 300000 |
| browserCount | Number | 10 |
| expectedTimeout | Number | 60000 |
| addedSlowmotion | Number | 1000 |
| initialPageloadDelay | Number | 1 |
| noHeadless | Boolean | false |
| debug | Boolean | false |
| auth | Boolean | false |
