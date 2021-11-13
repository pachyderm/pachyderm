## pachctl run pfs-load-test

Run a PFS load test.

### Synopsis

Run a PFS load test.

```
pachctl run pfs-load-test <spec-file> [flags]
```

### Examples

```

Specification:

-- CommitSpec --

count: int
operations: [ OperationSpec ]
validator:
  frequency: FrequencySpec
throughput:
  limit: int
  prob: int [0, 100]
cancel:
  prob: int [0, 100]
fileSources: [ FileSourceSpec ]

-- OperationSpec --

count: int
operation:
  putFile: PutFileSpec
  deleteFile: DeleteFileSpec
  prob: int [0, 100]

-- PutFileSpec --

count: int 
file: [ FileSpec ]

-- FileSpec --

source: string
prob: int [0, 100]

-- DeleteFileSpec --

count: int 
directoryProb: int [0, 100]

-- FrequencySpec --

count: int
prob: int [0, 100]

-- FileSourceSpec --

name: string 
random: RandomFileSourceSpec

-- RandomFileSourceSpec --

incrementPath: bool
directory: RandomDirectorySpec
size: [ SizeSpec ]

-- RandomDirectorySpec --

depth: int
run: int

-- SizeSpec --

min: int
max: int
prob: int [0, 100]

Example: 

count: 5
operations:
  - count: 5
    operation:
      - putFile:
          files:
            count: 5
            file:
              - source: "random"
                prob: 100
        prob: 70 
      - deleteFile:
          count: 5
          directoryProb: 20 
        prob: 30 
validator: {}
fileSources:
  - name: "random"
    random:
      directory:
        depth: 3
        run: 3
      size:
        - min: 1000
          max: 10000
          prob: 30 
        - min: 10000
          max: 100000
          prob: 30 
        - min: 1000000
          max: 10000000
          prob: 30 
        - min: 10000000
          max: 100000000
          prob: 10 

```

### Options

```
  -b, --branch string   The branch to use for generating the load.
  -h, --help            help for pfs-load-test
  -s, --seed int        The seed to use for generating the load.
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

