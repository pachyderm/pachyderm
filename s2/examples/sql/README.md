# s2 example

This example builds off of s2 and [gorm](http://gorm.io/) to provide an S3-like API, with objects stored in an in-memory sqlite instance. Note that data is not persisted, and a few simplifications have been made to make the example more manageable.

## Tests

To run tests, start the server with `make run`. Then, in a separate shell, run `make test`.
