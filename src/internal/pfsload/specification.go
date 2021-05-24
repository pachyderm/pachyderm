package pfsload

const LoadSpecification string = `
Specification:

-- CommitSpec --

count: int
operations: [ OperationSpec ]
validator:
  frequency: FrequencySpec
throughput:
  limit: int
  prob: float [0, 1]
cancel:
  prob: float [0, 1]
fileSources: [ FileSourceSpec ]

-- OperationSpec --

count: int
operation:
  putFile: PutFileSpec
  deleteFile: DeleteFileSpec
  prob: float [0, 1]

-- PutFileSpec --

count: int 
file: [ FileSpec ]

-- FileSpec --

source: string
prob: float [0, 1]

-- DeleteFileSpec --

count: int 
directoryProb: float [0, 1]

-- FrequencySpec --

count: int
prob: int

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
prob: float [0, 1]

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
                prob: 1
        prob: 0.7
      - deleteFile:
          count: 5
          directoryProb: 0.2
        prob: 0.3
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
          prob: 0.3
        - min: 10000
          max: 100000
          prob: 0.3
        - min: 1000000
          max: 10000000
          prob: 0.3
        - min: 10000000
          max: 100000000
          prob: 0.1
`
