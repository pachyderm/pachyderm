Composing Pipelines
===================

In any reasonably complex analysis isn't just going be computed in a single pipeline, but instead a chain of pipelines. We often refer to chains of pipelines as dependency graph or DAG (directed acyclic graph).  

Multiple Inputs
---------------

A pipeline is allowed to have multiple inputs.  The important thing to understand is what happens when a new commit comes into one of the input repos.  In short, a pipeline processes the **cross product** of its inputs.  We will use an example to illustrate.

Consider a pipeline that has two input repos: `foo` and `bar`.  `foo` uses the `file/incremental` method and `bar` uses the `reduce` method.  Now let's say that the following events occur:

```
1. PUT /file-1 in commit1 in foo -- no jobs triggered
2. PUT /file-a in commit1 in bar -- triggers job1
3. PUT /file-2 in commit2 in foo -- triggers job2
4. PUT /file-b in commit2 in bar -- triggers job3
```

The first time the pipeline is triggered will be when the second event completes.  This is because we need data in both repos before we can run the pipeline.

Here is a breakdown of the files that each job sees:

```
job1:
    /pfs/foo/file-1
    /pfs/bar/file-a

job2:
    /pfs/foo/file-2
    /pfs/bar/file-a

job3:
    /pfs/foo/file-1
    /pfs/foo/file-2
    /pfs/bar/file-a
    /pfs/bar/file-b
```

`job1` sees `/pfs/foo/file-1` and `/pfs/bar/file-a` because those are the only files available.

`job2` sees `/pfs/foo/file-2` and `/pfs/bar/file-a` because it's triggered by commit2 in `foo`, and `foo` uses an incremental input method (`file/incremental`).

`job3` sees all the files because it's triggered by commit2 in `bar`, and `bar` uses a non-incremental input method (`reduce`).

Transitive Reductions
---------------------


