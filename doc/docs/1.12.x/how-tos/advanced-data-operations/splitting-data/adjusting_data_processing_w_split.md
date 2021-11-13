# Adjusting Data Processing by Splitting Data

!!! info
    Before you read this section, make sure that you understand
    the concepts described in [File](../../../concepts/data-concepts/file.md),
    [Glob Pattern](../../../concepts/pipeline-concepts/datum/glob-pattern.md),
    [Pipeline Specification](../../../reference/pipeline_spec.md), and
    [Developer Workflow](../../developer-workflow/index.md).

Unlike source code version-control systems, such as Git, that mostly
store and version text files, Pachyderm does not perform intra-file
diffing. This means that if you have a 10,000 lines CSV file and
change a single line in that file, a pipeline that is subscribed
to the repository where that file is located processes the whole file.
You can adjust this behavior by splitting your file upon loading
into chunks.

Pachyderm applies diffing at per file level.
Therefore, if one bit of a file changes,
Pachyderm identifies that change as a new file.
Similarly, Pachyderm can only distribute computation
at the level of a single file. If your data is stored in
one large file, it can only be processed by a single worker, which
might affect performance.

To optimize performance and processing time, you might want to
break up large files into smaller chunks while Pachyderm uploads
data to an input repository. For simple data types, you can
run the `pachctl put file` command with the `--split` flag. For
more complex splitting pattern, such as when you work with `.avro`
or other binary formats, you need to manually split your data
either at ingest or by configuring splitting in a Pachyderm
pipeline.

## Using Split and Target-File Flags

For common file types that are often used in data science, such as CSV,
line-delimited text files, JavaScript Object Notation (JSON) files,
Pachyderm includes the `--split`, `--target-file-bytes`, and
`--target-file-datums` flags.

!!! note
    In this section, a chunk of data is called a *split-file*.

| Flag              | Description                                           |
| ----------------- | ----------------------------------------------------- |
| `--split`         | Divides a file into chunks based on a *record*, such as newlines <br> in a line-delimited files or by JSON object for JSON files. <br> The `--split` flag takes one of the following arguments— `line`, <br> `json`, or `sql`. For example, `--split line` ensures that Pachyderm <br> only breaks up a file on a newline boundaries and not in the <br> middle of a line. |
| `--target-file-bytes` |  This flag must be used with the `--split` <br> flag. The `--target-file-bytes` flag fills each of the split-files with data up to <br> the specified number of bytes, splitting on the nearest <br>record boundary. For example, you have a line-delimited file <br>of 50 lines, with each line having about 20 bytes. If you run the <br> `--split lines --target-file-bytes 100` command, you see the input <br>file split into about 10 files and each file has about 5 lines. Each <br>split-file’s size hovers above the target value of 100 bytes, <br>not going below 100 bytes until the last split-file, which might <br>be less than 100 bytes. |
| `--target-file-datums` | This flag must be used with the `--split` <br> flag. The `--target-file-datums` attempts to fill each split-file <br>with the specified number of datums. If you run `--split lines --target-file-datums 2` on the line-delimited 100-line file mentioned above, you see the file split into 50 split-files and each file has 2 lines. |


If you specify both `--target-file-datums` and `--target-file-bytes` flags,
Pachyderm creates split-files until it hits one of the
constraints.

!!! note "See also:"
    [Splitting Data for Distributed Processing](../splitting/#ingesting-postgressql-data)

### Example: Splitting a File

In this example, you have a 50-line file called `my-data.txt`.
You create a repository named `line-data` and load
`my-data.txt` into that repository. Then, you can analyze
how the data is split in the repository.

To complete this example, perform the following steps:

1. Create a file with fifty lines named `my-data.txt`. You can
   add random lines, such as numbers from one to fifty, or US states,
   or anything else.

      **Examples:**

      ```shell
      Zero
      One
      Two
      Three
      ...
      Fifty
      ```

1. Create a repository called `line-data`:

      ```shell
      pachctl create repo line-data
      pachctl put file line-data@master -f my-data.txt --split line
      ```

1. List the filesystem objects in the repository:

      ```shell
      pachctl list file line-data@master
      ```

      **System Response:**

      ```shell
      NAME         TYPE  SIZE
      /my-data.txt dir   1.071KiB
      ```

      The `pachctl list file` command shows
      that the line-oriented file `my-data.txt`
      that was uploaded has been transformed into a
      directory that includes the chunks of the original
      `my-data.txt` file. Each chunk is put into a split-file
      and given a 16-character filename, left-padded with 0.
      Pachyderm numbers each filename sequentially in hexadecimal. We
      modify the command to list the contents of “my-data.txt”, and the output
      reveals the naming structure.

1. List the files in the `my-data.txt` directory:

      ```shell
      pachctl list file line-data@master my-data.txt
      ```

      **System Response:**

      ```shell
      NAME                          TYPE  SIZE
      /my-data.txt/0000000000000000 file  21B
      /my-data.txt/0000000000000001 file  22B
      /my-data.txt/0000000000000002 file  24B
      /my-data.txt/0000000000000003 file  21B
      ...
      NAME                          TYPE  SIZE
      /my-data.txt/0000000000000031 file  22B
      ```

## Example: Appending to files with `-–split`

Combining `--split` with the default *append* behavior of
`pachctl put file` enables flexible and scalable processing of
record-oriented file data from external, legacy systems.

Pachyderm ensures that only the newly added data gets processed when
you append to an existing file by using `--split`. Pachyderm
optimizes storage utilization by automatically deduplicating each
split-file. If you split a large file
with many duplicate lines or objects with identical hashes
might use less space in PFS than it does as
a single file outside of PFS.

To complete this example, follow these steps:

1. Create a file `count.txt` with the following lines:

      ```shell
      Zero
      One
      Two
      Three
      Four
      ```

1. Put the `count.txt` file into a Pachyderm repository called `raw_data`:

      ```shell
      pachctl put file -f count.txt raw_data@master --split line
      ```

      This command splits the `count.txt` file by line and creates
      a separate file with one line in each file. Also, this operation
      creates five datums that are processed by the
      pipelines that use this repository as input.

1. View the repository contents:

      ```shell
      pachctl list file raw_data@master
      ```

      **System Response:**

      ```shell
      NAME       TYPE SIZE
      /count.txt dir  24B
      ```

      Pachyderm created a directory called `count.txt`.

1. View the contents of the `count.txt` directory:

      ```shell
      pachctl list file raw_data@master:count.txt
      ```

      **System Response:**

      ```shell
      NAME                        TYPE SIZE
      /count.txt/0000000000000000 file 4B
      /count.txt/0000000000000001 file 4B
      /count.txt/0000000000000002 file 6B
      /count.txt/0000000000000003 file 5B
      /count.txt/0000000000000004 file 5B
      ```

      In the output above, you can see that Pachyderm created five split-files
      from the original `count.txt` file. Each file has one line from the
      original `count.txt`. You can check the contents of each file by
      running the `pachctl get file` command. For example, to get
      the contents of `count.txt/0000000000000000`, run the following
      command:

      ```shell
      pachctl get file raw_data@master:count.txt/0000000000000000
      ```

      **System Response:**

      ```shell
      Zero
      ```

      This operation creates five datums that are processed by the
      pipelines that use this repository as input.

1. Create a one-line file called `more-count.txt` with the
   following content:

      ```shell
      Five
      ```

1. Load this file into Pachyderm by appending it to the `count.txt` file:

      ```shell
      pachctl put file raw_data@master:count.txt -f more-count.txt --split line
      ```

    If you do not specify `--split` flag while appending to
    a file that was previously split, Pachyderm displays the following
    error message:

     ```shell
     could not put file at "/count.txt"; a file of type directory is already there
     ```

1. Verify that another file was added:

      ```shell
      pachctl list file raw_data@master:count.txt
      ```

      **System Response:**

      ```shell
      NAME                        TYPE SIZE
      /count.txt/0000000000000000 file 4B
      /count.txt/0000000000000001 file 4B
      /count.txt/0000000000000002 file 6B
      /count.txt/0000000000000003 file 5B
      /count.txt/0000000000000004 file 5B
      /count.txt/0000000000000005 file 4B
      ```

      The `/count.txt/0000000000000005` file was added to the input
      repository. Pachyderm considers
      this new file as a separate datum. Therefore, pipelines process
      only that datum instead of all the chunks of `count.txt`.

1. Get the contents of the `/count.txt/0000000000000005` file:

      ```
      pachctl get file raw_data@master:count.txt/0000000000000005
      ```

      **System Response:**

      ```shell
      Five
      ```

## Example: Overwriting Files with `–-split`

The behavior of Pachyderm when a file loaded with ``--split`` is
overwritten is simple to explain but subtle in its implications.
Most importantly, it can have major implications when new rows
are inserted within the file as opposed to just being appended to the end.
The loaded file is split into those sequentially-named files,
as shown above. If any of those resulting
split-files hashes differently than the one it is replacing, it
causes the Pachyderm Pipeline System to process that data.
This can have significant consequences for downstream processing.

To complete this example, follow these steps:

1. Create a file `count.txt` with the following lines:

      ```shell
      One
      Two
      Three
      Four
      Five
      ```

1. Put the file into a Pachyderm repository called `raw_data`:

      ```shell
      pachctl put file -f count.txt raw_data@master --split line
      ```

      This command splits the `count.txt` file by line and creates
      a separate file with one line in each file. Also, this operation
      creates five datums that are processed by the
      pipelines that use this repository as input.

1. View the repository contents:

      ```shell
      pachctl list file raw_data@master
      ```

      **System Response:**

      ```shell
      NAME       TYPE SIZE
      /count.txt dir  24B
      ```

      Pachyderm created a directory called `count.txt`.

1. View the contents of the `count.txt` directory:

      ```shell
      pachctl list file raw_data@master:count.txt
      ```

      **System Response:**

      ```shell
      NAME                        TYPE SIZE
      /count.txt/0000000000000000 file 4B
      /count.txt/0000000000000001 file 4B
      /count.txt/0000000000000002 file 6B
      /count.txt/0000000000000003 file 5B
      /count.txt/0000000000000004 file 5B
      ```

      In the output above, you can see that Pachyderm created five split-files
      from the original `count.txt` file. Each file has one line from the
      original `count.txt` file. You can check the contents of each file by
      running the `pachctl get file` command. For example, to get
      the contents of `count.txt/0000000000000000`, run the following
      command:

      ```shell
      pachctl get file raw_data@master:count.txt/0000000000000000
      ```

      **System Response:**

      ```shell
      One
      ```

1. In your local directory, modify the original `count.txt` file by
   inserting the word *Zero* on the first line:

      ```shell
      Zero
      One
      Two
      Three
      Four
      Five
      ```

1. Upload the updated `count.txt` file into the raw_data repository
   by using the `--split` and `--overwrite` flags:

      ```shell
      pachctl put file -f count.txt raw_data@master:count.txt --split line --overwrite
      ```

      Because Pachyderm takes the file name into account when hashing
      data for a pipeline, it considers every single split-file as new,
      and the pipelines that use this repository as input process all
      six datums.

1. List the files in the directory:

      ```shell
      pachctl list file raw_data@master:count.txt
      ```

      **System Response:**

      ```shell
      NAME                        TYPE SIZE
      /count.txt/0000000000000000 file 5B
      /count.txt/0000000000000001 file 4B
      /count.txt/0000000000000002 file 4B
      /count.txt/0000000000000003 file 6B
      /count.txt/0000000000000004 file 5B
      /count.txt/0000000000000005 file 5B
      ```

      The `/count.txt/0000000000000000` file now has the newly added `Zero` line.
      To verify the contents of the file, run:

      ```shell
      pachctl get file raw_data@master:count.txt/0000000000000000
      ```

      **System Response:**

      ```shell
      Zero
      ```

!!! note "See Also:"
    [Splitting Data](../splitting/)
