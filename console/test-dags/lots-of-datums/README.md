### Build the image

you can tag and upload to your own docker repo or just use the image in `pipeline.json`

```
docker build -t peterpachyderm/datum:v1 .
docker push peterpachyderm/datum:v1
```

### Pach setup

```
pachctl create repo data
pachctl create pipeline -f pipeline.json
```

Generate random files using `generateData.js`

```
mkdir -p data
node generateData.js 100
pachctl put file -r data@master:/ -f data

pachctl list file generate-datums.meta@master:/pfs
```
