/*
Package index provides access to files through multilevel indexes.

A multilevel index contains one or more levels where the lowest level contains
file type index entries which directly reference file content and the above
levels contain range type index entries which directly reference a range of
index entries and indirectly reference ranges of file content. Multilevel
indexes are created using writers which provide functionality for creating
indexes for new files or creating indexes based on other indexes (rooted by
file or range type indexes). Reading a multilevel index requires setting up a
reader which provide various indexing strategies and filters.
*/
package index
