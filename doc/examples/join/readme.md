# Premise

The goal for this compound pipeline is to compute the best selling author and best selling publisher.

## Run It Yourself

To run the whole thing you can just run `make all` if you already have a pachctl
cluster setup. This will allow you to follow along with the commands we issue at
each stage of this guide.

## Schema

| Sales|
|----|---|
| Sale ID | Book ID |

| Authors |
|----|---|
| ID | Name | Publisher ID |

| Books |
|----|---|
| ID | Title | Author ID | Price |

## Basic Design

Basically we want to join over these tables ...

Sales x Books (via book ID)

Books x Authors (via author ID)

So that we have all of the columns we need to aggregate by author ID / publisher ID ... sum over price ... and sort the results.

## Stages

### Data

Each table is stored under a directory in a single file. Later on we'll add some more sales data.


```
$pc list-file data master /authors
NAME           TYPE SIZE 
/authors/1.csv file 55B  
$pc list-file data master /books
NAME         TYPE SIZE     
/books/1.csv file 5.947KiB 
$pc list-file data master /sales
NAME         TYPE SIZE 
/sales/1.csv file 586B 

```


### Map

In this stage, we want to divide the data up per row. So instead of a single file with all the rows, we want one row per file. This will allow us to operate on rows in subsequent steps.


#### Input

| Input Repo | Glob Pattern |
|---|---|
| data | `/` |

The input to the `map` pipeline is the whole (`/`) of the `data` repository. The code itself will walk over the directories (each table) and split out the files into individual rows.

#### Output

```
$pc list-file map master
NAME    TYPE SIZE     
authors dir  40B      
books   dir  5.792KiB 
sales   dir  482B     
$pc list-file map master authors
NAME           TYPE SIZE 
/authors/1.csv file 7B   
/authors/2.csv file 9B   
/authors/3.csv file 7B   
/authors/4.csv file 8B   
/authors/5.csv file 9B   
$pc get-file map master /authors/1.csv 
1,bob,0
```

Here we see that each table has been broken out so that each row is an individual file. In each case, the filename is the ID of the row as defined by the table.

### Join

In this stage, we need to stitch these data sets together. We do this on a per row basis. In particular, we want to join sales to books via the `bookID`, and books to authors by the `authorID`. 

#### Input

To accomplish this, let's take a look at our input:

```
  "input": {
        "cross": [
            {
                "atom":{
                    "repo": "map",
                    "name": "sales",
                    "glob": "/sales/*"
                }
            },
            {
                "atom":{
                    "repo": "map",
                    "name": "authors",
                    "glob": "/authors/*"
                }
            },
            {
                "atom":{
                    "repo": "map",
                    "name": "books",
                    "glob": "/books/*"
                }
            }
        ]
    }
```

We're taking a cross product of our inputs. [You can refer here to more docs on joins and crosses](http://docs.pachyderm.io/en/latest/cookbook/combining.html)

Let's break this down a bit. The first input is just: 

```
"atom":{
    "repo": "map",
    "name": "sales",
    "glob": "/sales/*"
}
```

Which means we're taking the sales data from output of the previous map pipeline. We specify a glob of `/sales/*` which means each file under the sales directory will be processed seperately.

How many files is that? Well if we look ...

```
18-10-18[15:20:47]:join:0$pc list-file map master /sales
NAME          TYPE SIZE 
/sales/1.csv  file 4B   
...
```

We see that it's a lot. In fact it's 10:

```
$pc list-file map master /sales | wc -l
11
# minus the headers, this is 10 values
```

Similarly for the other two inputs, we're selecting each of the files separately. Then we take the cross product. That means every combination of files will be seen in some datum. For instance, some datum will see:

```
/pfs/authors/authors/3.csv
/pfs/books/books/6.csv
/pfs/sales/sales/4.csv
```

Which means there will be as many combinations as there are # of authors times # of books times # of sales. How many combinations exactly? Well we need to know the number of each of these.

```
$pc list-file map master /sales | wc -l
11
# minus the headers, this is 10 values
$pc list-file map master /authors | wc -l
7
# so 6 total
$pc list-file map master /books | wc -l
11
# and 10 total
```

To summarize ...


| Input Repo | Glob Pattern | Datum Count |
|---|---|---|
| data | `/authors/authors/*` | 6 |
| data | `/sales/sales/*` | 10 |
| data | `/books/books/*` | 10 |

So that's 600 combinations in total. This is something to be aware of and keep an eye on when designing your pipelines. Depending on your workload too many datums will introduce too much overhead.

#### Processing

The code reads each of the single rows presented. That means there's one row each describing an author, book, and a sale. Then the code checks if their join keys are equal.

```
# join.rb

# This is the line that checks the keys match ...
if sale[1] == book[0] && book[2] == author[0]
	f = File.open(File.join(OUTPUT_DIR, sale[0] + ".csv"), "w") 
	cols, data = join(sale, book, author)
	f << cols.join(",") + "\n"
	f << data.join(",")
end
```

So let's say this particular datum sees the combination:

```
/pfs/authors/authors/2.csv
/pfs/books/books/7.csv
/pfs/sales/sales/4.csv
```

If sale number 4 has a book ID that matches book number 7 AND book number 7 has an author ID
that matches author number 2 ... then we write out the full row.

But what if they don't match? That's ok. We don't need to write anything out
since the cross product guarantees _some_ datum will see the right combination.

#### Parallelism

We have 600 datums to process here. If we process these sequentially, even if
they're fast, this will quickly add up.

By default we have this pipeline running with a single worker. As is, it'll take
about 5 minutes to complete. 

But the good news is we can distribute this join. Since each datum can be processed in any order, we can spread the work across workers. 

If we supply the following config in our join pipeline:

```
"parallelism_spec": {
  "constant": 10
}
```

We'll see the job complete in about 30 seconds.

#### Output

We output the joined rows, with the filename based on the sale ID:

```
$pc list-file join master
NAME   TYPE SIZE 
1.csv  file 58B  
10.csv file 59B  
11.csv file 59B  
12.csv file 59B  
```

### Shuffle

Now that we have all the raw rows joined the right way, we need to aggregate
them in a way that makes sense. In particular, since we want to know who the
best selling author and best selling publishers are ... we want to group these
rows by `authorID` and `publisherID`.

#### Input

| Input Repo | Glob Pattern |
|---|---|
| join | `/*` |

Which means we process each joined row by itself. We could crank up parallelism
if this was a lot of rows, but since it's only 100, we'll leave parallelism at
the default of a single worker.

#### Processing

To do that, we take each row, and write it to 2 places. To
`/pfs/out/authors/{authorID}.csv` and `/pfs/out/publishers/{publisherID}.csv`.
We make sure we append to these files, as many subsequent datums will be writing
to files of the same name. This way (by appending) the aggregate result will be
all the rows that contain the given ID.

#### Output

```
8-10-18[15:37:48]:join:0$pc list-file shuffle master
NAME       TYPE SIZE     
authors    dir  1.293KiB 
publishers dir  1.293KiB 
18-10-18[15:52:18]:join:0$pc list-file shuffle master /authors
NAME           TYPE SIZE 
/authors/1.csv file 239B 
/authors/2.csv file 253B 
/authors/3.csv file 237B 
/authors/4.csv file 357B 
/authors/5.csv file 238B 
18-10-18[15:52:21]:join:0$pc get-file shuffle master /authors/2.csv
15,48,2,1,3.36
23,45,2,1,3.40
24,84,2,1,8.72
58,73,2,1,9.59
53,84,2,1,8.72
51,74,2,1,2.53
66,45,2,1,3.40
61,89,2,1,0.03
65,88,2,1,3.60
62,44,2,1,1.76
71,48,2,1,3.36
68,79,2,1,7.41
92,47,2,1,0.98
96,88,2,1,3.60
99,44,2,1,1.76
1,44,2,1,1.76
4,73,2,1,9.59
```

Here we see all the joined rows for author number 2.


### Reduce

Now that we have the rows binned by author or publisher ID, we can calculate the
total sales for each author/publisher ... then sort the result.

#### Input

| Input Repo | Glob Pattern | Datum Count |
|---|---|---|
| shuffle | `/*` | 2 |
| data | `/` | 1 |

And since we're doing the cross, this means that there will be 2 datums -- one for authors, one for publishers.

e.g. One datum will see:

```
/pfs/shuffle/authors
```

While the other will see:

```
/pfs/shuffle/publishers
```

#### Processing

The code will read each of the individual files under the table given, aggregate
the sum, sort the results - e.g. tuples of [author, total sales] - and write
them to a single file.

We also cross reference the top result (the author/publisher w the most sales)
w the original tables to lookup the name. These are emitted in the output as
well

#### Parallelism

We set the parallelism to 2 to demonstrate that we can do these reduces per
table separately.

#### Output

We see the raw rankings for authors / publishers. And we see the top ranked
results.

```
$pc list-file reduce master
NAME       TYPE SIZE 
rankings   dir  58B  
top_seller dir  18B  
$pc list-file reduce master /rankings
NAME                     TYPE SIZE 
/rankings/authors.csv    file 42B  
/rankings/publishers.csv file 16B  
$pc get-file reduce master /rankings/authors.csv
4,42.81999999999999
2,24.79
1,14.94
3,7.4
$pc list-file reduce master /top_seller
NAME                       TYPE SIZE 
/top_seller/authors.txt    file 5B   
/top_seller/publishers.txt file 13B  
$pc get-file reduce master /top_seller/authors.txt
emma
$pc get-file reduce master /top_seller/publishers.txt
random house
```

## Add More Data

You can run `make new-sales` to add a bit more data to the sales table ... and
watch it percolate through the pipelines.


