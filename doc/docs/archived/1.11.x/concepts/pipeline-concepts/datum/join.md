# Join

A join is a special type of pipeline input that enables you to combine
files that reside in separate Pachyderm repositories and match a
particular naming pattern. The join operator must be used in combination
with a glob pattern that reflects a specific naming convention.

By analogy, a Pachyderm join is similar to a database *equi-join*,
or *inner join* operation, but it matches on file paths
only, not the contents of the files.

Unlike the [cross input](../datum/cross-union.md), which creates datums
from every combination of files in each input repository, joins only create
datums where there is a *match*. You can use joins to combine data from
different Pachyderm repositories and ensure that only specific files from
each repo are processed together.

When you configure a join input, you must specify a glob pattern that
includes a capture group. The capture group defines the specific string in
the file path that is used to match files in other joined repos.
Capture groups work analogously to the [regex capture group](https://www.regular-expressions.info/refcapture.html).
You define the capture group inside parenthesis. Capture groups are numbered
from left to right and can also be nested within each other. Numbering for
nested capture groups is based on their opening parenthesis.

Below you can find a few examples of applying a glob pattern with a capture
group to a file path. For example, if you have the following file path:

```shell
/foo/bar-123/ABC.txt
```

The following glob patterns in a joint input create the
following capture groups:

| Regular expression  | Capture groups           |
| ------------------- | ------------------------ |
| `/(*)`              | `foo`                    |
| `/*/bar-(*)`        | `123`                    |
| `/(*)/*/(??)*.txt`  | Capture group 1: `foo`, capture group 2: `AB`. |
| `/*/(bar-(123))/*`  | Capture group 1: `bar-123`, capture group 2: `123`. |


Also, joins require you to specify a [replacement group](https://www.regular-expressions.info/replacebackref.html)
in the `join_on` parameter to define which capture groups you want to try
to match.

For example, `$1` indicates that you want Pachyderm to match based on
capture group `1`. Similarly, `$2` matches the capture group `2`.
`$1$2` means that it must match both capture groups `1` and `2`.

If Pachyderm does not find any matching files, you get a zero-datum job.

You can test your glob pattern and capture groups by using the
`pachctl glob file` command as described in
[Glob Pattern](../../datum/glob-pattern/#test-a-glob-pattern).

## Example

For example, you have two repositories. One with sensor readings
and the other with parameters. The repositories have the following
structures:

* `readings` repo:

   ```shell
   ├── ID1234
       ├── file1.txt
       ├── file2.txt
       ├── file3.txt
       ├── file4.txt
       ├── file5.txt
   ```

* `parameters` repo:

   ```shell
   ├── file1.txt
   ├── file2.txt
   ├── file3.txt
   ├── file4.txt
   ├── file5.txt
   ├── file6.txt
   ├── file7.txt
   ├── file8.txt
   ```

Pachyderm runs your code only on the pairs of files that match
the glob pattern and capture groups.

The following example shows how you can use joins to group
matching IDs:

```json
 {
   "pipeline": {
     "name": "joins"
   },
   "input": {
     "join": [
       {
         "pfs": {
           "repo": "readings",
           "branch": "master",
           "glob": "/*/(*).txt",
           "join_on": "$1"
         }
       },
      {
        "pfs": {
          "repo": "parameters",
          "branch": "master",
          "glob": "/(*).txt",
          "join_on": "$1"
        }
      }
    ]
  },
  "transform": {
     "cmd": [ "python3", "/joins.py"],
     "image": "joins-example"
   }
 }
```

The glob pattern for the `readings` repository, `/*/(*).txt`, indicates all
matching files in the `ID` sub-directory. In the `parameters` repository,
the glob pattern `/(*).txt` selects all the matching files in the root
directory.
All files with indices from `1` to `5` match. The files
with indices from `6` to `8` do not match. Therefore, you only get five
datums for this job.

To experiment further, see the full [joins example](https://github.com/pachyderm/pachyderm/tree/1.13.x/examples/joins).
