This is a small library for working with modified
[Merkle Trees](https://en.wikipedia.org/wiki/Merkle_tree). We store one of these
data structures in block storage (e.g. S3) for each PFS commit, so that we know,
with each subsequent commit, what files changed and need to be reprocessed by
any pipelines.
