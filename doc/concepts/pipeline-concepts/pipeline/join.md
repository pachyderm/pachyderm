# Join

A join is a special type of pipeline input that enables you
to combine files that reside in separate Pachyderm repositories
and match a particular naming pattern. The join operator must
be used in combination with a glob pattern that reflects a
specific naming convention.

A Pachyderm join is similar to a database equi-join operation that
combines columns from one or more tables into its own dataset
if the column names match.

Similarly, joins help you consolidate results with matching names from
multiple repositories in a single dataset. For example, you collect
data from multiple sources and need to consolidate the results
that match a specific parameter. These results can be readings
from a sensor, a geographical location, or something similar.

When you configure a join input, you must specify a glob pattern
that defines a capture group. A capture group is the group of
files selected for further processing. In the glob parameter,
a capturing glob must be enclosed in parenthesis. For example,
if you have a file named `test`, you could configure the
following globs for it:

* `(t..t)`, `join_on = $1$2` translates to `es`.
* `(t..t)`, `join_on = $2$1` translates to `se`.

If Pachyderm does not find matching files, this pipeline run
results in a zero-datum job.

To test your glob pattern, use the `pachctl glob file` command
as described in [Glob Pattern](../datum/glob-pattern.html#test-glob-patterns.html).

## Example

For example, you have two repositories. One with sensor readings
and the other with parameters. The repositories have the following
structures:

* `readings`:

   ```bash
   ├── ID1234
       ├── file1.txt
       ├── file2.txt
       ├── file3.txt
       ├── file4.txt
       ├── file5.txt
   ```

* `parameters`:

   ```bash
   ├── file1.txt
   ├── file2.txt
   ├── file3.txt
   ├── file4.txt
   ├── file5.txt
   ├── file6.txt
   ├── file7.txt
   ├── file8.txt
   ```

Pachyderm runs your code on the files that match the
glob pattern. For example, your code can multiply lines in matching files to
create a joined file with the product of both files or perform other
operations as needed.

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

The glob pattern for the `readings` repository, `/*/(*)`, indicates all
matching files in the `ID` sub-directory. In the `parameters` repository,
the glob pattern `/(*)` selects all the matching files in the root directory.
All files with indices from `1` to `5` match. The files
with indices from `6` to `8` do not match.

For more information, see the [Joins example](https://github.com/pachyderm/pachyderm/tree/master/examples/join).
