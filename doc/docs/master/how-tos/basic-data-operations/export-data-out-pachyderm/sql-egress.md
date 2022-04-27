
# Egress To An SQL Database

!!! Warning
    SQL Egress is an [experimental feature](../../../contributing/supported-releases/#experimental){target=_blank}.

Pachyderm already implements [egress to object storage](../export-data-egress){target=_blank} as an optional egress field in the pipeline specification. 
Similarly, our **SQL egress** lets you seamlessly export data from a Pachyderm-powered pipeline output repo to a SQL database. Specifically, we help you connect to a remote database of your choice and push the content of a CSV or a JSON file to tables of your choice, matching the column names and casting their content into their SQL datatype.

## Use SQL Egress

To egress data from the output commit of a pipeline to a SQL database, you will need to:

 1. *On your cluster*: 

    [Create a **secret**](#create-a-secret) containing your database password. 

 1. *In the Specification file of your egress pipeline*:

    Provide the [**connection string**](#update-your-pipeline-spec) to the database and choose the **format of the files (CVS or JSON)** containing the data to insert.

 1. *In your user code*: 

    Write your data files in [a root directory named after the table](#3-in-your-user-code-write-your-data-to-directories-named-after-each-table) you want to insert them into. 
    You can have multiple directories.

### 1- Create a Secret 

Create a **secret** containing your database password in the field `PACHYDERM_SQL_PASSWORD`. This secret is identical to the database secret of Pachyderm SQL Ingest. Refer to the SQL Ingest page for instructions on [how to create your secret](../../sql-ingest/#database-secret){target=_blank}.

### 2- Update your Pipeline Spec

Append an egress section to your pipeline specification file, then fill in:

- the `URL`: the connection string to your database. Its format is identical to the [url in the SQL Ingest](../../sql-ingest/#database-connection-url){target=_blank}.
- the `file_format` type: CSV or JSON.

!!! Example
        ```json
        {
            "pipeline": {
                "name": "SQLegress"
            },
            "input": {
                "pfs": {
                    "repo": "input_repo",
                    "glob": "/"
                }
            },
            "transform": {
            ...
            },
            "egress": {
                "URL": "snowflake://pachyderm@CDKSGMN-UN89564/PACH_TEST/PUBLIC?warehouse=COMPUTE_WH",
                "sql_options": {
                    "file_format": {
                        "type": "CSV"
                    },
                    "secret": {
                        "k8s_secret": "databasesecretname",
                        "key": "PACHYDERM_SQL_PASSWORD",
                        "env_var": "PACHYDERM_SQL_PASSWORD"
                    }
                }
            }
        }
        ```

!!! Warning 
    In the case of a JSON format, you will additionally need to provide a declarative list of column names. 
    ```json
    "file_format": {
            "type": "JSON",
            "options": {
                "json_field_names": ["ID", "NAME", "MY_DATE"]
            }
    },
    ```

### 3- In your User Code, Write Your Data to Directories Named After Each Table
 
The user code of your pipeline determines what data should be egressed and to which tables. 
Data that the pipeline writes to the output repo is interpreted as tables corresponding to directories. 

**Each top-level directory is named after the table you want to egress its content to**. Each root directory is walked; all of the files reachable in the walk are parsed in a given format indicated by an egress parameter, e.g., JSON, CSV, then inserted in the corresponding table. 

!!! Warning
     - Files in the root produce an error as they do not correspond to a table.
     - The directory structure below the top level does not matter.  The first directory in the path is the table; everything else is walked until a file is found.  All the data in those files is inserted into the table.

- For CSV files:

    The order of the values in each line of the csv MUST match the order of the columns in the schema of your table.
    
    !!! Example 
            ```
            "1","Tim","2017-03-12T21:51:45Z","true"
            "12","Tom","2017-07-25T21:51:45Z","true"
            "33","Tam","2017-01-01T21:51:45Z","false"
            "54","Pach","2017-05-15T21:51:45Z","true"
            ```

- For JSON files:

    The column names are part of the file.

    !!! Example
            ```
            {"ID": "11","NAME": "I will get there","MY_DATE": "2017-07-25T21:51:45Z"}
            ```


!!! Note 
    - Pachyderm queries the schema of the database before the insertion so the data are parsed into their corresponding SQL data types.
    - The table you insert values into must pre-exist.
    - Each insertion created a new row in your table (no upsert).


## Troubleshooting

You have a pipeline running but do not see any update in your database? 

Check your logs:

```shell
pachctl list pipeline
pachctl logs -p <your-pipeline-name> --master
```


  

