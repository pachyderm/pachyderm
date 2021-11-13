# Cross and Union Inputs

<!---This section needs to be made more clear. There is a lot of information
that I would say describes the things you can do with a cross or union pipeline
but does not really have a good and clear explanation of what they are -->

Pachyderm enables you to combine multiple
input repositories in a single pipeline by using the `union` and
`cross` operators in the pipeline specification.

If you are familiar with [Set theory](https://en.wikipedia.org/wiki/Set_theory),
you can think of union as a *disjoint union binary operator* and cross as a
*cartesian product binary operator*. However, if you are unfamiliar with these
concepts, it is still easy to understand how cross and union work in Pachyderm.

This section describes how to use `cross` and `union` in your pipelines and how you
can optimize your code when you work with them.

## Union Input

The union input combines each of the datums in the input repos as one
set of datums.
The number of datums that are processed is the sum of all the
datums in each repo.

For example, you have two input repos, `A` and `B`. Each of these
repositories contain three files with the following names.

Repository `A` has the following structure:

```shell
A
├── 1.txt
├── 2.txt
└── 3.txt
```

Repository `B` has the following structure:

```shell
B
├── 4.txt
├── 5.txt
└── 6.txt
```

If you want your pipeline to process each file independently as a
separate datum, use a glob pattern of `/*`. Each
glob is applied to each input independently. The input section
in the pipeline spec might have the following structure:

```shell
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
directory, so three datums from each input. Therefore, the union of `A` and `B`
has six datums in total.
Your pipeline processes the following datums without any specific order:

```shell
/pfs/A/1.txt
/pfs/A/2.txt
/pfs/A/3.txt
/pfs/B/4.txt
/pfs/B/5.txt
/pfs/B/6.txt
```

!!! note
    Each datum in a pipeline is processed independently by a single
    execution of your code. In this example, your code runs six times, and
    each datum is available to it one at a time. For example, your code
    processes `pfs/A/1.txt` in one of the runs and `pfs/B/5.txt` in a
    different run, and so on. In a union, two or more datums are never
    available to your code at the same time. You can simplify
    your union code by using the `name` property as described below.

### Simplifying the Union Pipelines Code

In the example above, your code needs to read into the `pfs/A`
_or_ `pfs/B` directory because only one of them is present in any given datum.
To simplify your code, you can add the `name` field to the `pfs` object and
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

```shell
/pfs/C/1.txt  # from A
/pfs/C/2.txt  # from A
/pfs/C/3.txt  # from A
/pfs/C/4.txt  # from B
/pfs/C/5.txt  # from B
/pfs/C/6.txt  # from B
```

## Cross Input

In a cross input, Pachyderm exposes every combination of datums,
or a cross-product, from each of your input repositories to your code
in a single run.
In other words, a cross input pairs every datum in one repository with
each datum in another, creating sets of datums. Your transformation
code is provided one of these sets at the time to process.

For example, you have repositories `A` and `B` with three datums, each
with the following structure:

!!! note
    For this example, the glob pattern is set to `/*`.

Repository `A` has three files at the top level:

```shell
A
├── 1.txt
├── 2.txt
└── 3.txt
```

Repository `B` has three files at the top level:

```shell
B
├── 4.txt
├── 5.txt
└── 6.txt
```

Because you have three datums in each repo, Pachyderm exposes
a total of nine combinations of datums to your code.

!!! important
    In cross pipelines, both `pfs/A` and `pfs/B`
    directories are visible during each code run.

```shell
Run 1: /pfs/A/1.txt
       /pfs/B/1.txt

Run 2: /pfs/A/1.txt
       /pfs/B/2.txt
...

Run 9: /pfs/A/3.txt
       /pfs/B/3.txt
```

!!! note
    In cross inputs, if you use the `name` field, your two
    inputs cannot have the same name. This could cause file system collisions.

!!! note "See Also:"

- [Cross Input](../../../../reference/pipeline_spec/#cross-input)
- [Union Input](../../../../reference/pipeline_spec/#union-input)
- [Distributed hyperparameter tuning](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/ml/hyperparameter)

<!-- Add a link to an interactive tutorial when it's ready-->
