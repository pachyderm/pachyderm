# JSON to SQL

This directory contains a small go script that consumes single-layer .json files from stdin and outputs them as SQL INSERT statements. Each field in each record of the input is converted to a column name, and the value of that field is written to that column.

The binary takes a single argument: the name of the table that the data should be written to. For example:

```
$ go build to_sql.go && cat test.json | ./to_sql cars
INSERT INTO cars (year, make, model) VALUES (2005, Toyota, Corolla);
INSERT INTO cars (model, year, make) VALUES (Civic, 1998, Honda);
INSERT INTO cars (make, model, year) VALUES (Tesla, Roadster, 2008);
INSERT INTO cars (make, model, year) VALUES (Bugatti, Chiron, 2016);
INSERT INTO cars (make, model, year) VALUES (Dodge, Viper, 2015);
```
