/*
Package chunk provides access to data through content-addressed chunks.

A chunk is the basic unit of storage for data. Chunks are identified by their
content-address, which is a hash of the data stored in the chunk. There are two
mechanisms for uploading chunks: uploader and batcher. The uploader is intended
for medium / large data entries since it performs content-defined chunking on
the data and stores each data entry in its own set of chunks. The batcher is
intended for small data entries since it batches multiple data entries into
larger chunks. The result of each of these upload methods is a list of data
references for each data entry. These data references can be used to access the
corresponding data through readers.
*/
package chunk
