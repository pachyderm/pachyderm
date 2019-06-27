# Datum

Datum is a unit of computation that Pachyderm processes and that
defines how data is written to output files. Depending on
the anticipated result, you can represent each dataset as a smaller
subset of data. You can configure a whole
input repository to be one datum, each filesystem object in
the root directory to be a separate datum,
filesystem objects in the subdirectories to be separate datums,
and so on. Datums affect how Pachyderm distributes proccessing workloads
among underlying compute nodes and are instrumental in optimizing
your configuration for best performance.

Pachyderm takes each datum and processes them in isolation on one of
the undrlying worker nodes. Then, results from all the nodes are combined
togethter and represented as one or multiple datums in the output repository.
All these properties can be configured through the user-defined variables in
the pipeline specification.

To understand how datum affects data processing in Pachyderm, you need to
understand the following subconcepts:

* Relationship between datums
* Glob Pattern
* Parallelism

## Relationship between datums

Pachyderm pipelines take a datum or datums from the source, processes
it or them, and places the result either as a single datum or multiple
datums to the output repository. You can configure how you want your
data to be processed, either as a single datum or as multiple datums.
You can also configure how you want the data to be uploaded to the output
repository - either as one datum, or file, or as multiple datums, or files.
This concept is called relationship between datums.

Relationships between datums boil down to the following categories:

| Relationship       | Description                                                                                      |
| ------------------ | ------------------------------------------------------------------------------------------------ |
| One-to-one (1-1)   | A unique datum in the input repository maps to a unique datum <br> in  the output repository.    |
| One-to-many (1-M)  | A unique datum in the input repository maps to multiple datums <br>    in the output repository. |
| Many-to-many (M-M) | Many unique datums in the input repository map to many datums  <br>    in the output repository. |

### One-to-one relationship

In one-to-one relationship, Pachyderm takes one datum from the input
repository, transforms the data, and places the result as a single datum
in the output repository.

The [OpenCV example] illustrates the one-to-one relationship.

### One-to-many relationship

In one-to-many relationship, Pachyderm takes one datum from the input
repository, transforms the data, and places the result as a single datum
in the output repository.

The [OpenCV example] illustrates the one-to-one relationship.

### One-to-many relationship

In one-to-many relationship, Pachyderm takes one datum from the input
repository and maps it to many datums in the output repository. For example,
the image in the OpenCV example can be split into many tiles for further
analysis. Each tile represents a unique datum.

### Many-to-many relationship

Most of the pipelines that are created in Pachyderm have many-to-many
relationships. The pipeline takes many unique datums from the input
repository or repositories and writes to many output files in the
output repository. The output files might not be unique across
datums. Therefore, the results from processing many datums can
be represented in a single file. Pachyderm can merge all the
results into a single file. When you run `pachctl list jobs`,
you can sometimes see the `merging` status in the output of the
command which means that Pachyderm optimizes the result by
presenting it in a single file.

The [Wordcount example](https://github.com/pachyderm/pachyderm/tree/master/examples/word_count)
illustrates the many-to-many relationship.

For example, you want to analyze which words are unique in the first paragraph
from *Moby Dick by Herman Melville* stored as `md.txt` and
in the first paragraph of *A Tale of Two Cities by Charles Dickens* stored as
`ttc.md`.

After running the pipeline, Pachyderm collects the following results:

* The words `was`, `best`, `worst`, `wisdom`, and `foolishness` are unique
to *A Tale of Two Cities*. Therefore, each file name after the related word
contains only the output from processing `ttc.txt`.

* The words `Ishmael` and `money` are unique to *Moby Dick*. Therefore, the
files named after the related word only contain output from processing
`md.txt`.

* The words `it`, `the`, and `of` are present in both extracts.
The pipeline outputs the results to the word file in each container
and Pachyderm merges the results.



**Note:** Data is processed randomly, without any particular order.

## Glob Pattern


