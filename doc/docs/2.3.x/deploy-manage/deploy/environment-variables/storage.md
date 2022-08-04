# Configure Storage Variables

## Variables 

| Variable | Required | Type | Default | Description |
|---|---|---|---|---|
| STORAGE_MEMORY_THRESHOLD | No | int64 |  | The threshold for storage memory. |
| STORAGE_COMPACTION_SHARD_SIZE_THRESHOLD | No | int64 |  | The threshold for the total size of all files in a shard. |
| STORAGE_COMPACTION_SHARD_COUNT_THRESHOLD | No | int64 |  | The threshold for the total count of all files in a shard. |
| STORAGE_LEVEL_FACTOR | No | int64 |  |  |
| STORAGE_UPLOAD_CONCURRENCY_LIMIT | No | int | 100 |  |
| STORAGE_PUT_FILE_CONCURRENCY_LIMIT | No | int | 100 |  |
| STORAGE_GC_PERIOD | No | int64 | 60 | The number of seconds between PFS's garbage collection cycles; garbage collection is disabled when this setting is < `0`. |
| STORAGE_CHUNK_GC_PERIOD | No | int64 | 60 | The number of seconds between chunk garbage colletion cycles; chunk garbage collection is disabled when this setting is < `0`. |
| STORAGE_COMPACTION_MAX_FANIN | No | int | 10 |  |
| STORAGE_FILESETS_MAX_OPEN | No | int | 50 |  |
| STORAGE_DISK_CACHE_SIZE | No | int | 100 |  |
| STORAGE_MEMORY_CACHE_SIZE | No | int | 100 |  |