# Join

A join is a special type of pipeline input that enables you to combine
files that reside in separate Pachyderm repositories and match a
particular naming pattern. The join operator must be used in combination
with a glob pattern that reflects a specific naming convention.

By analogy, a Pachyderm join is similar to a database *equi-join*,
or *inner left join* operation, but it matches on file paths
only, not the contents of the files.

Unlike the [cross input](cross-input.md) which creates datums from every
combination of files in each input repository, joins only create datums
where there is a *match*. This can be incredibly useful for combining
data from different repositories or ensuring that only specific files from each repo are processed together.

When you configure a join input, you must specify a glob pattern that also includes a capture group. The capture group defines the specific string in the file path that is used to match against files in the other joined repos. Capture groups work analagously to [regex capture group](https://www.regular-expressions.info/refcapture.html) -- you define the capture group with parenthesis. Capture groups are numbered from left to right and can also be nested within each other (numbering for nested capture groups are based on their opening parenthesis).

Here are a few examples of applying a glob pattern with a capture group to file path:

```
sample file path:
/foo/bar-123/ABC.txt

/(*)                --> c1="foo"

/*/bar-(*)          --> c1="123"

/(*)/*/(??)*.txt    --> c1="foo", c2="AB"

/*/(bar-(123))/*    --> c1="bar-123", c2="123"
```

Joins also require a `join_on` parameter, also called [replacement group](https://www.regular-expressions.info/replacebackref.html) to define which capture groups you want to try to match.  

For example, `$1` would indicate that you want pachyderm to match based on capture group 1. `$2` would similarly mean match capture group 2. `$1$2` would mean that it must match both capture groups 1 and 2.

If Pachyderm does not find any matching files, you will get a zero-datum job.

You can test your glob pattern and capture groups using the `pachctl glob file` command as described in [Glob Pattern](../datum/glob-pattern.html#test-glob-patterns.html).

## Example

For example, you have two repositories. One with sensor readings
and the other with parameters. The repositories have the following
structures:

* `readings` repo:

   ```bash
   ├── ID1234
       ├── file1.txt
       ├── file2.txt
       ├── file3.txt
       ├── file4.txt
       ├── file5.txt
   ```

* `parameters` repo:

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

Pachyderm runs your code on only the pairs of files that match the glob pattern and capture groups.

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
with indices from `6` to `8` do not match. Therefore, you will only get 5 datums for this job.

To experiment further, see the full [joins example](https://github.com/pachyderm/pachyderm/tree/master/examples/join).
