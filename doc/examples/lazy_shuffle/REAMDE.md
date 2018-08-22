# Creating a lazy shuffle pipeline

This example demonstrates how lazy shuffle pipeline i.e. a pipeline that shuffles, combines files without downloading/uploading can be created. For more information [see](https://pachyderm.readthedocs.io/en/latest/managing_pachyderm/data_management.html)

## Create fruits input repo
```bash
pachctl create-repo fruits
pachctl put-file fruits master -f mango.jpeg
pachctl put-file fruits master -f apple.jpeg
```

## Create pricing input repo
```bash
pachctl create-repo pricing
pachctl put-file pricing master -f mango.json
pachctl put-file pricing master -f apple.json
```


## Create lazy shuffle pipeline
```bash
pachctl create-pipeline -f lazy_shuffle.json
```


## Results

### List-job
`pachctl  list-job` indicates no data download or upload was performed

| ID                               | OUTPUT COMMIT                                 | STARTED        | DURATION  | RESTART | PROGRESS  | DL | UL | STATE   |
|----------------------------------|-----------------------------------------------|----------------|-----------|---------|-----------|----|----|---------|
| 60617fd06155451d8358cc714bf9b670 | lazy_shuffle/f56e97fa9e234eb6ad902640d4fba2ac | 10 seconds ago | 4 seconds | 0       | 4 + 0 / 4 | 0B | 0B | success |



### Output files:
`pachctl list-file lazy_shuffle master "*"` will show shuffled file:

| NAME             | TYPE | SIZE     |
|------------------|------|----------|
| /mango/cost.json | file | 22B      |
| /mango/img.jpeg  | file | 7.029KiB |
| /apple/cost.json | file | 23B      |
| /apple/img.jpeg  | file | 4.978KiB |
