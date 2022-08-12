# Group Input

A group is a special type of pipeline input that enables you to aggregate
files that reside in one or separate Pachyderm repositories and match a
particular naming pattern. The group operator must be used in combination
with a glob pattern that reflects a specific naming convention.

By analogy, a Pachyderm group is similar to a database *group-by*,
but it matches on file paths only, not the content of the files.

Unlike the [join](../datum/join.md) datum that will always contain a single match (even partial) from each input repo,
**a group creates one datum for each set of matching files accross its input repos**.
You can use group to aggregate data that is not adequately captured by your directory structure 
or to control the granularity of your datums through file name-matching. 


When you configure a group input, you must specify a glob pattern that
includes a capture group. The capture group defines the specific string in
the file path that is used to match files in other grouped repos.
Capture groups work analogously to the [regex capture group](https://www.regular-expressions.info/refcapture.html){target=_blank}.
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


Also, groups require you to specify a [replacement group](https://www.regular-expressions.info/replacebackref.html){target=_blank}
in the `group_by` parameter to define which capture groups you want to try
to match.

For example, `$1` indicates that you want Pachyderm to match based on
capture group `1`. Similarly, `$2` matches the capture group `2`.
`$1$2` means that it must match both capture groups `1` and `2`.

If Pachyderm does not find any matching files, you get a zero-datum job.

You can test your glob pattern and capture groups by using the
`pachctl list datum -f <your_pipeline_spec.json>` command as described in
[List Datum](../../datum/glob-pattern/#test-your-datums).

!!! Important "Useful"
    The content of the capture group defined in the `group_by` parameter is available to your pipeline's code in an environment variable: `PACH_DATUM_<input.name>_GROUP_BY`.
## Example

For example, a repository `labresults` contains the lab results of patients. 
The files at the root of your repository have the following naming convention. You want to group your lab results by patientID.

* `labresults` repo:

   ```shell
   ├── LIPID-patientID1-labID1.txt (1)
   ├── LIPID-patientID2-labID1.txt (2)
   ├── LIPID-patientID1-labID2.txt (3)
   ├── LIPID-patientID3-labID3.txt (4)
   ├── LIPID-patientID1-labID3.txt (5)
   ├── LIPID-patientID2-labID3.txt (6)
   ```

Pachyderm runs your code on the set of files that match
the glob pattern and capture groups.

The following example shows how you can use group to aggregate all the lab results of each patient.

```json
 {
   "pipeline": {
     "name": "group"
   },
   "input": {
     "group": [
      {
        "pfs": {
          "repo": "labresults",
          "branch": "master",
          "glob": "/*-(*)-lab*.txt",
          "group_by": "$1"
        }
      }
    ]
  },
   "transform": {
      "cmd": [ "bash" ],
      "stdin": [ "wc" ,"-l" ,"/pfs/labresults/*" ]
      }
  }
 }
```

The glob pattern for the `labresults` repository, `/*-(*)-lab*.txt`, selects all files with a patientID match in the root directory.

The pipeline will process 3 datums for this job. 

- all files containing `patientID1` (1, 3, 5) are grouped in one datum, 
- a second datum will be made of (2, 6) for `patientID2`
- and a third with (4) for `patientID3`

The `pachctl list datum -f <your_pipeline_spec.json>` command is a useful tool to check your datums: 

```code
ID FILES                                                                                                                                                                                                                        STATUS TIME
-  labresults@722665ed49474db0aab5cbe4d8a20ff8:/LIPID-patientID1-labID1.txt, labresults@722665ed49474db0aab5cbe4d8a20ff8:/LIPID-patientID1-labID3.txt, labresults@722665ed49474db0aab5cbe4d8a20ff8:/LIPID-patientID1-labID2.txt -      -
-  labresults@722665ed49474db0aab5cbe4d8a20ff8:/LIPID-patientID2-labID1.txt, labresults@722665ed49474db0aab5cbe4d8a20ff8:/LIPID-patientID2-labID3.txt                                                                           -      -
-  labresults@722665ed49474db0aab5cbe4d8a20ff8:/LIPID-patientID3-labID3.txt
```

To experiment further, see the full [group example](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/examples/group){target=_blank}.

