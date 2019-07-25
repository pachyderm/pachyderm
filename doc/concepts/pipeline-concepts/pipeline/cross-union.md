# Cross and Union Inputs

<!---This section needs to be made more clear. There is a lot of information
that I would say describes the things you can do with a cross or union pipeline
but does not really have a good and clear explanation of what they are -->

Pachyderm enables you to combine multiple PFS inputs, or multiple
input repositories, in a single pipeline by using the `union` and
`cross` operators in the pipeline specification. Such pipelines
are called *union pipeline or input* and *cross pipeline or input*.

If you are familiar with the [Set theory](https://en.wikipedia.org/wiki/Set_theory),
you can think of union as a *disjoint union binary operator* and of cross as a
`cartesian product binary operator`. However, if you are unfamiliar with these
concepts, it is still easy to understand how cross and union work in Pachyderm.

This section describes what union and cross pipelines are and how you
can optimize your code when you work with them.

## Union Input

The union input combines each of the datums in the input repos as one
set of datums.
The number of datums that are processed is the sum of all the
datums in each repo.

For example, you have two input repos, `A` and `B`. Each of these
repositories contain three files with the following identical names.

Repository `A` has the following structure:

```bash
A
├── 1.txt
├── 2.txt
└── 3.txt
```

Repository `B` has the following structure:

```bash
B
├── 1.txt
├── 2.txt
└── 3.txt
```

Although these files have identical names, each file has different content.
Therefore, each file has a different hash. If you combine them in a
pipeline, the `input` object in the pipeline spec might have the following
structure:

```bash
"input": {
    "union": [
        {
            "pfs": {
                "glob": "/*",
                "repo": "A"
            }
        },
        {
            "pfs": {
                "glob": "/*",
                "repo": "B"
            }
        }
    ]
}
```

In this example, each Pachyderm repository has those three files in the root
directory, then you have six datums in total, which is the sum of the number
of input files.
Your pipeline processes the following datums without any specific order:

```bash
/pfs/A/1.txt
/pfs/A/2.txt
/pfs/A/3.txt
/pfs/B/1.txt
/pfs/B/2.txt
/pfs/B/3.txt
```

### Optimizing Union Pipelines

To simplify your code, you can add a `name` field to the `pfs` object and
give the same name to each of the input repos. For example, you can add, the
`name` field with the value `C` to the input repositories `A` and `B`:

```
"input": {
    "union": [
        {
            "pfs": {
                "name": "C",
                "glob": "/*",
                "repo": "A"
            }
        },
        {
            "pfs": {
                "name": "C",
                "glob": "/*",
                "repo": "B"
            }
        }
    ]
}
```

Then, in the pipeline, all datums appear in the same directory.

```bash
/pfs/C/1.txt  # from A
/pfs/C/2.txt  # from A
/pfs/C/3.txt  # from A
/pfs/C/1.txt  # from B
/pfs/C/2.txt  # from B
/pfs/C/3.txt  # from B
```

## Cross Input

A cross input, instead of appending datums from one repository to datums
in the other repository,
multiplies each datum from one repository with each datum from another
repository. In other words, a cross input
combines all the datums and provides them to the pipeline
that uses this combination as input and treats each combination
as a separate datum.

For example, you have repositories `A` and `B` with the following
structures:

**Note:** For this example, the glob pattern is set to `/*`

Repository `A` has the following structure:

```bash
A
├── 1.txt
├── 2.txt
└── 3.txt
```

Repository `B` has the following structure:

```bash
B
├── 1.txt
├── 2.txt
└── 3.txt
```


Although the files have identical names, each file has different content.
To create a combination of each file, the pipeline needs to process nine
datums without any specific order:

**Important:** In cross pipelines, both directories are visible during
processing.

```bash
/pfs/A/1.txt    /pfs/A/1.txt    /pfs/A/1.txt
/pfs/B/1.txt    /pfs/B/2.txt    /pfs/B/3.txt

/pfs/A/2.txt    /pfs/A/2.txt    /pfs/A/1.txt
/pfs/B/1.txt    /pfs/B/2.txt    /pfs/B/2.txt

/pfs/A/3.txt    /pfs/A/3.txt    /pfs/A/1.txt
/pfs/B/1.txt    /pfs/B/2.txt    /pfs/B/3.txt
```

In cross inputs, you cannot specify the same `name` to the combined
input repositories to avoid name collisions. Both directories are
visible during processing.

**See Also:**

- [Cross Input](../../../reference/pipeline_spec.html#cross-input)
- [Union Input](../../../reference/pipeline_spec.html#union-input)
- [Combining/Merging/Joining Data](../../../cookbook/combining.html#combining-merging-joining-data)
- [Distributed hyperparameter tuning](https://github.com/pachyderm/pachyderm/tree/master/examples/ml/hyperparameter)

<!-- Add a link to an interactive tutorial when it's ready-->
