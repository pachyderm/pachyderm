# GoRethink - RethinkDB Driver for Go 

[![GitHub tag](https://img.shields.io/github/tag/dancannon/gorethink.svg?style=flat)](https://github.com/dancannon/gorethink/releases)
[![GoDoc](https://godoc.org/github.com/dancannon/gorethink?status.png)](https://godoc.org/github.com/dancannon/gorethink)
[![build status](https://img.shields.io/travis/dancannon/gorethink/master.svg "build status")](https://travis-ci.org/dancannon/gorethink) 

[Go](http://golang.org/) driver for [RethinkDB](http://www.rethinkdb.com/) 

![GoRethink Logo](https://raw.github.com/wiki/dancannon/gorethink/gopher-and-thinker-s.png "Golang Gopher and RethinkDB Thinker")

Current version: v2.0.4 (RethinkDB v2.3)

Please note that this version of the driver only supports versions of RethinkDB using the v0.4 protocol (any versions of the driver older than RethinkDB 2.0 will not work).

If you need any help you can find me on the [RethinkDB slack](http://slack.rethinkdb.com/) in the #gorethink channel.

## Installation

```
go get gopkg.in/dancannon/gorethink.v2
```

(Or v1)

```sh
go get gopkg.in/dancannon/gorethink.v1
```

## Connection

### Basic Connection

Setting up a basic connection with RethinkDB is simple:

```go
import (
    r "github.com/dancannon/gorethink"
    "log"
)

var session *r.Session

session, err := r.Connect(r.ConnectOpts{
    Address: "localhost:28015",
})
if err != nil {
    log.Fatalln(err.Error())
}

```
See the [documentation](http://godoc.org/github.com/dancannon/gorethink#Connect) for a list of supported arguments to Connect().

### Connection Pool

The driver uses a connection pool at all times, by default it creates and frees connections automatically. It's safe for concurrent use by multiple goroutines.

To configure the connection pool `MaxIdle`, `MaxOpen` and `Timeout` can be specified during connection. If you wish to change the value of `MaxIdle` or `MaxOpen` during runtime then the functions `SetMaxIdleConns` and `SetMaxOpenConns` can be used.

```go
var session *r.Session

session, err := r.Connect(r.ConnectOpts{
    Address: "localhost:28015",
    Database: "test",
    MaxIdle: 10,
    MaxOpen: 10,
})
if err != nil {
    log.Fatalln(err.Error())
}

session.SetMaxOpenConns(5)
```

### Connect to a cluster

To connect to a RethinkDB cluster which has multiple nodes you can use the following syntax. When connecting to a cluster with multiple nodes queries will be distributed between these nodes.

```go
var session *r.Session

session, err := r.Connect(r.ConnectOpts{
    Addresses: []string{"localhost:28015", "localhost:28016"},
    Database: "test",
    AuthKey:  "14daak1cad13dj",
    DiscoverHosts: true,
})
if err != nil {
    log.Fatalln(err.Error())
}
```

When `DiscoverHosts` is true any nodes are added to the cluster after the initial connection then the new node will be added to the pool of available nodes used by GoRethink. Unfortunately the canonical address of each server in the cluster **MUST** be set as otherwise clients will try to connect to the database nodes locally. For more information about how to set a RethinkDB servers canonical address set this page http://www.rethinkdb.com/docs/config-file/.

## User Authentication

To login with a username and password you should first create a user, this can be done by writing to the `users` system table and then grant that user access to any tables or databases they need access to. This queries can also be executed in the RethinkDB admin console.

```go
err := r.DB("rethinkdb").Table("users").Insert(map[string]string{
    "id": "john",
    "password": "p455w0rd",
}).Exec(session)
...
err = r.DB("blog").Table("posts").Grant("john", map[string]bool{
    "read": true,
    "write": true,
}).Exec(session)
...
```

Finally the username and password should be passed to `Connect` when creating your session, for example:

```go
session, err := r.Connect(r.ConnectOpts{
    Address: "localhost:28015",
    Database: "blog",
    Username: "john",
    Password: "p455w0rd",
})
```

Please note that `DiscoverHosts` will not work with user authentication at this time due to the fact that RethinkDB restricts access to the required system tables.

## Query Functions

This library is based on the official drivers so the code on the [API](http://www.rethinkdb.com/api/) page should require very few changes to work.

To view full documentation for the query functions check the [API reference](https://github.com/dancannon/gorethink/wiki/Go-ReQL-command-reference) or [GoDoc](http://godoc.org/github.com/dancannon/gorethink#Term)

Slice Expr Example
```go
r.Expr([]interface{}{1, 2, 3, 4, 5}).Run(session)
```
Map Expr Example
```go
r.Expr(map[string]interface{}{"a": 1, "b": 2, "c": 3}).Run(session)
```
Get Example
```go
r.DB("database").Table("table").Get("GUID").Run(session)
```
Map Example (Func)
```go
r.Expr([]interface{}{1, 2, 3, 4, 5}).Map(func (row Term) interface{} {
    return row.Add(1)
}).Run(session)
```
Map Example (Implicit)
```go
r.Expr([]interface{}{1, 2, 3, 4, 5}).Map(r.Row.Add(1)).Run(session)
```
Between (Optional Args) Example
```go
r.DB("database").Table("table").Between(1, 10, r.BetweenOpts{
    Index: "num",
    RightBound: "closed",
}).Run(session)
```

### Optional Arguments

As shown above in the Between example optional arguments are passed to the function as a struct. Each function that has optional arguments as a related struct. This structs are named in the format FunctionNameOpts, for example BetweenOpts is the related struct for Between.

## Results

Different result types are returned depending on what function is used to execute the query.

- `Run` returns a cursor which can be used to view all rows returned.
- `RunWrite` returns a WriteResponse and should be used for queries such as Insert, Update, etc...
- `Exec` sends a query to the server and closes the connection immediately after reading the response from the database. If you do not wish to wait for the response then you can set the `NoReply` flag.

Example:

```go
res, err := r.DB("database").Table("tablename").Get(key).Run(session)
if err != nil {
    // error
}
defer res.Close() // Always ensure you close the cursor to ensure connections are not leaked
```

Cursors have a number of methods available for accessing the query results

- `Next` retrieves the next document from the result set, blocking if necessary.
- `All` retrieves all documents from the result set into the provided slice.
- `One` retrieves the first document from the result set.

Examples:

```go
var row interface{}
for res.Next(&row) {
    // Do something with row
}
if res.Err() != nil {
    // error
}
```

```go
var rows []interface{}
err := res.All(&rows)
if err != nil {
    // error
}
```

```go
var row interface{}
err := res.One(&row)
if err == r.ErrEmptyResult {
    // row not found
}
if err != nil {
    // error
}
```

## Encoding/Decoding
When passing structs to Expr(And functions that use Expr such as Insert, Update) the structs are encoded into a map before being sent to the server. Each exported field is added to the map unless

  - the field's tag is "-", or
  - the field is empty and its tag specifies the "omitempty" option.

Each fields default name in the map is the field name but can be specified in the struct field's tag value. The "gorethink" key in
the struct field's tag value is the key name, followed by an optional comma
and options. Examples:

```go
// Field is ignored by this package.
Field int `gorethink:"-"`
// Field appears as key "myName".
Field int `gorethink:"myName"`
// Field appears as key "myName" and
// the field is omitted from the object if its value is empty,
// as defined above.
Field int `gorethink:"myName,omitempty"`
// Field appears as key "Field" (the default), but
// the field is skipped if empty.
// Note the leading comma.
Field int `gorethink:",omitempty"`
```

**NOTE:** It is strongly recommended that struct tags are used to explicitly define the mapping between your Go type and how the data is stored by RethinkDB. This is especially important when using an `Id` field as by default RethinkDB will create a field named `id` as the primary key (note that the RethinkDB field is lowercase but the Go version starts with a capital letter).

When encoding maps with non-string keys the key values are automatically converted to strings where possible, however it is recommended that you use strings where possible (for example `map[string]T`).

If you wish to use the `json` tags for GoRethink then you can call `SetTags("gorethink", "json")` when starting your program, this will cause GoRethink to check for `json` tags after checking for `gorethink` tags. By default this feature is disabled. This function will also let you support any other tags, the driver will check for tags in the same order as the parameters.

### References

Sometimes you may want to use a Go struct that references a document in another table, instead of creating a new struct which is just used when writing to RethinkDB you can annotate your struct with the reference tag option. This will tell GoRethink that when encoding your data it should "pluck" the ID field from the nested document and use that instead.

This is all quite complicated so hopefully this example should help. First lets assume you have two types `Author` and `Book` and you want to insert a new book into your database however you dont want to include the entire author struct in the books table. As you can see the `Author` field in the `Book` struct has some extra tags, firstly we have added the `reference` tag option which tells GoRethink to pluck a field from the `Author` struct instead of inserting the whole author document. We also have the `gorethink_ref` tag which tells GoRethink to look for the `id` field in the `Author` document, without this tag GoRethink would instead look for the `author_id` field.

```go
type Author struct {
    ID      string  `gorethink:"id,omitempty"`
    Name    string  `gorethink:"name"`
}

type Book struct {
    ID      string  `gorethink:"id,omitempty"`
    Title   string  `gorethink:"title"`
    Author  Author `gorethink:"author_id,reference" gorethink_ref:"id"`
}
```

The resulting data in RethinkDB should look something like this:

```json
{
    "author_id": "c2182a10-6b9d-4ea1-a70c-d6649bb5f8d7",
    "id":  "eeb006d6-7fec-46c8-9d29-45b83f07ca14",
    "title":  "The Hobbit"
}
```

If you wanted to read back the book with the author included then you could run the following GoRethink query:

```go
r.Table("books").Get("1").Merge(func(p r.Term) interface{} {
    return map[string]interface{}{
        "author_id": r.Table("authors").Get(p.Field("author_id")),
    }
}).Run(session)
```

## Logging

By default the driver logs errors when it fails to connect to the database. If you would like more verbose error logging you can call `r.SetVerbose(true)`.

Alternatively if you wish to modify the logging behaviour you can modify the logger provided by `github.com/Sirupsen/logrus`. For example the following code completely disable the logger:

```go
r.Log.Out = ioutil.Discard
```

## Benchmarks

Everyone wants their project's benchmarks to be speedy. And while we know that rethinkDb and the gorethink driver are quite fast, our primary goal is for our benchmarks to be correct. They are designed to give you, the user, an accurate picture of writes per second (w/s). If you come up with a accurate test that meets this aim, submit a pull request please. 

Thanks to @jaredfolkins for the contribution.

| Type    |  Value   |
| --- | --- |
| **Model Name** | MacBook Pro |
| **Model Identifier** | MacBookPro11,3 |
| **Processor Name** | Intel Core i7 | 
| **Processor Speed** | 2.3 GHz | 
| **Number of Processors** | 1 |
| **Total Number of Cores** | 4 |
| **L2 Cache (per Core)** | 256 KB | 
| **L3 Cache** | 6 MB | 
| **Memory** | 16 GB |

```bash
BenchmarkBatch200RandomWrites                20                              557227775                     ns/op
BenchmarkBatch200RandomWritesParallel10      30                              354465417                     ns/op
BenchmarkBatch200SoftRandomWritesParallel10  100                             761639276                     ns/op
BenchmarkRandomWrites                        100                             10456580                      ns/op
BenchmarkRandomWritesParallel10              1000                            1614175                       ns/op
BenchmarkRandomSoftWrites                    3000                            589660                        ns/op
BenchmarkRandomSoftWritesParallel10          10000                           247588                        ns/op
BenchmarkSequentialWrites                    50                              24408285                      ns/op
BenchmarkSequentialWritesParallel10          1000                            1755373                       ns/op
BenchmarkSequentialSoftWrites                3000                            631211                        ns/op
BenchmarkSequentialSoftWritesParallel10      10000                           263481                        ns/op
```

## Examples

Many functions have examples and are viewable in the godoc, alternatively view some more full features examples on the [wiki](https://github.com/dancannon/gorethink/wiki/Examples).

## Further reading

- [GoRethink Goes 1.0](https://www.compose.io/articles/gorethink-goes-1-0/)
- [Go, RethinkDB & Changefeeds](https://www.compose.io/articles/go-rethinkdb-and-changefeeds-part-1/)
- [Build an IRC bot in Go with RethinkDB changefeeds](http://rethinkdb.com/blog/go-irc-bot/)

## License

Copyright 2013 Daniel Cannon

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Donations

[![Donations](https://pledgie.com/campaigns/29517.png "Donations")](https://pledgie.com/campaigns/29517) 
