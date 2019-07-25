
# Relationship Between Datums

A Pachyderm pipeline takes a datum or datums from the source, processes
it or them, and places the result either as a single datum or multiple
datums to the output repository. You can configure how you want your
data to be processed, as a single datum or as multiple datums.
You can also configure how you want the data to be uploaded to the output
repository,either as one datum or file, or as multiple datums or files.
This concept is called the relationship between datums.

Relationships between datums boil down to the following categories:

| Relationship       | Description                                                                                   |
| ------------------ | --------------------------------------------------------------------------------------------- |
| One-to-one (1-1)   | A unique datum in the input repository maps to a unique datum <br> in the output repository.  |
| One-to-many (1-M)  | A unique datum in the input repository maps to multiple datums <br> in the output repository. |
| Many-to-many (M-M) | Many unique datums in the input repository map to many datums  <br> in the output repository. |

## One-to-one Relationship

In a one-to-one relationship, Pachyderm takes one datum from the input
repository, transforms the data, and places the result as a single datum
in the output repository.

<!--- Add a diagram -->

The [OpenCV example](../../../getting_started/beginner_tutorial.html)
illustrates the one-to-one relationship.

## One-to-many Relationship

In a one-to-many relationship, Pachyderm takes one datum from the input
repository and maps it to many datums in the output repository. For example,
the image in the OpenCV example can be split into many tiles for further
analysis. Each tile represents a unique datum.

<!--- Add a diagram -->

## Many-to-many Relationship

Most of the pipelines that users create in Pachyderm have many-to-many
relationships. The pipeline takes many unique datums from the input
repository or repositories and writes to many output files in the
output repository. The output files might not be unique across
datums. Therefore, the results from processing many datums can
be represented in a single file. Pachyderm can merge all the
results into a single file. When you run `pachctl list jobs`,
you can sometimes see the `merging` status in the output of the
command, which means that Pachyderm optimizes the result by
consolidating the processed datums in a single file.

The [Wordcount example](https://github.com/pachyderm/pachyderm/tree/master/examples/word_count)
illustrates the many-to-many relationship.

<!--- Add a diagram -->
